package eventsourcex

import (
	"context"
	"strings"
	"time"

	"github.com/altairsix/eventsource"
	"github.com/altairsix/pkg/action"
	"github.com/altairsix/pkg/action/ticker"
	"github.com/altairsix/pkg/checkpoint"
	"github.com/altairsix/pkg/tracer"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/opentracing/opentracing-go/log"
)

const (
	// DefaultCommitInterval the minimum amount of time that must pass between offset commits
	DefaultCommitInterval = time.Second * 3

	// DefaultPublishInterval the amount of time between checking the repository for updates
	DefaultPublishInterval = time.Minute
)

// Publisher publishes the record to a event bus
type Publisher interface {
	Publish(record eventsource.StreamRecord) error
}

// PublisherFunc provides a func wrapper to Publisher
type PublisherFunc func(record eventsource.StreamRecord) error

// Publish Implements the Publisher interface
func (fn PublisherFunc) Publish(record eventsource.StreamRecord) error { return fn(record) }

// Supervisor reads events from a StreamReader and supervisor them to a handler
type Supervisor interface {
	Check()
	Close() error
	Done() <-chan struct{}
}

type supervisor struct {
	ctx             context.Context
	cancel          func()
	done            chan struct{}
	check           chan struct{}
	segment         tracer.Segment
	r               eventsource.StreamReader
	h               Publisher
	cpKey           string
	cp              *checkpoint.CP
	interval        time.Duration
	offset          uint64
	offsetLoaded    bool
	committedOffset uint64
	committedAt     time.Time
	recordCount     int
}

// Close stops the worker process
func (s *supervisor) Close() error {
	s.cancel()
	<-s.done
	return nil
}

// Check request a check from the supervisor
func (s *supervisor) Check() {
	select {
	case s.check <- struct{}{}:
	default:
	}
}

// Done allows external tools to signal off of when the supervisor is done
func (s *supervisor) Done() <-chan struct{} {
	return s.done
}

func (s *supervisor) checkOnce() {
	segment, ctx := tracer.NewSegment(s.ctx, "supervisor:check_once")
	defer segment.Finish()

	if !s.offsetLoaded {
		v, err := s.cp.Load(ctx, s.cpKey)
		if err != nil {
			return
		}
		s.offset = v
		s.offsetLoaded = true
	}

	// read  events
	events, err := s.r.Read(ctx, s.offset+1, s.recordCount)
	if err != nil {
		return
	}

	// publish  events
	for _, event := range events {
		if err := s.h.Publish(event); err != nil {
			return
		}

		s.offset = event.Offset
	}

	// time to commit?
	if now := time.Now(); now.Sub(s.committedAt) > DefaultCommitInterval {
		if err := s.cp.Save(ctx, s.cpKey, s.offset); err == nil {
			s.committedAt = now
			s.committedOffset = s.offset
		}
	}
}

func (s *supervisor) listenAndPublish() {
	defer close(s.done)
	defer close(s.check)
	defer s.segment.Finish()

	s.segment.Info("supervisor:started", log.String("interval", s.interval.String()))

	timer := time.NewTicker(s.interval)
	defer timer.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-timer.C:
			s.checkOnce()
		case <-s.check:
			s.checkOnce()
		}
	}
}

// WithPublishEvents publishes received events to nats
func WithPublishEvents(fn Publisher, nc *nats.Conn, env, boundedContext string) PublisherFunc {
	rootSubject := StreamSubject(env, boundedContext) + "."

	return func(event eventsource.StreamRecord) error {
		if err := fn.Publish(event); err != nil {
			return err
		}

		subject := rootSubject + event.AggregateID
		go nc.Publish(subject, []byte(event.AggregateID))
		return nil
	}
}

// PublishStream reads from a stream and publishes
func PublishStream(ctx context.Context, h Publisher, r eventsource.StreamReader, cp *checkpoint.CP, env, bc string) Supervisor {
	cpKey := makeCheckpointKey(env, bc)
	segment, _ := tracer.NewSegment(ctx, "publish_stream", log.String("checkpoint", cpKey))

	child, cancel := context.WithCancel(ctx)
	p := &supervisor{
		ctx:         child,
		cancel:      cancel,
		done:        make(chan struct{}),
		check:       make(chan struct{}, 1),
		segment:     segment,
		r:           r,
		h:           h,
		cpKey:       cpKey,
		cp:          cp,
		interval:    DefaultPublishInterval,
		recordCount: 100,
	}

	go p.listenAndPublish()

	return p
}

type tracingPublisher struct {
	target  Supervisor
	segment tracer.Segment
}

// WithTraceReceiveNotices returns a Supervisor that ping when Check is invoked
func WithTraceReceiveNotices(s Supervisor, segment tracer.Segment) Supervisor {
	return &tracingPublisher{
		target:  s,
		segment: segment,
	}
}

func (s *tracingPublisher) Close() error          { return s.target.Close() }
func (s *tracingPublisher) Done() <-chan struct{} { return s.target.Done() }
func (s *tracingPublisher) Check() {
	s.segment.Info("eventsourcex:notice_received")
	s.target.Check()
}

// WithReceiveNotifications listens to nats for notices on the StreamSubject and prods the supervisor
func WithReceiveNotifications(s Supervisor, nc *nats.Conn, env, boundedContext string) Supervisor {
	go func() {
		subject := NoticesSubject(env, boundedContext)
		fn := func(m *nats.Msg) {
			s.Check()
		}

		var sub *nats.Subscription
		for {
			select {
			case <-s.Done():
				return
			default:
			}

			if v, err := nc.Subscribe(subject, fn); err == nil {
				sub = v
				break
			}
		}

		<-s.Done()
		sub.Unsubscribe()
	}()

	return s
}

// PublishStan publishes events to the nats stream identified with the env and boundedContext
func PublishStan(st stan.Conn, subject string) PublisherFunc {
	return func(event eventsource.StreamRecord) error {
		return st.Publish(subject, event.Data)
	}
}

// PublishStreamSingleton is similar to PublishStream except that there may be only one running in the environment
func PublishStreamSingleton(ctx context.Context, p Publisher, r eventsource.StreamReader, cp *checkpoint.CP, env, bc string, nc *nats.Conn) error {
	segment, ctx := tracer.NewSegment(ctx, "publish_stream")
	segment.SetBaggageItem("subject", StreamSubject(env, bc))
	defer segment.Finish()

	t := ticker.Nats(nc, makeTickerSubject(env, bc))
	fn := action.Singleton(t, func(ctx context.Context) error {
		h := WithPublishEvents(p, nc, env, bc)              // publish events to here
		supervisor := PublishStream(ctx, h, r, cp, env, bc) // go!
		if env == "local" {
			supervisor = WithTraceReceiveNotices(supervisor, segment)
		}
		supervisor = WithReceiveNotifications(supervisor, nc, env, bc) // ping the supervisor when events received
		<-supervisor.Done()                                            // wait until done
		return nil
	})

	return fn(ctx)
}

func makeCheckpointKey(env, bc string) string {
	return "stan:" + env + "." + bc
}

func makeTickerSubject(env, bc string, args ...string) string {
	key := "ticker:" + env + "." + bc
	if len(args) > 0 {
		key += "." + strings.Join(args, ".")
	}
	return key
}
