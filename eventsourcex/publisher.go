package eventsourcex

import (
	"context"
	"fmt"
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
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

const (
	// DefaultCommitInterval the minimum amount of time that must pass between offset commits
	DefaultCommitInterval = time.Second * 3

	// DefaultPublishInterval the amount of time between checking the repository for updates
	DefaultPublishInterval = time.Minute
)

// Publisher reads events from a StreamReader and publisher them to a handler
type Publisher interface {
	Check()
	Close() error
	Done() <-chan struct{}
}

type publisher struct {
	ctx             context.Context
	cancel          func()
	done            chan struct{}
	check           chan struct{}
	segment         tracer.Segment
	r               eventsource.StreamReader
	h               func(eventsource.StreamRecord) error
	cpKey           string
	cp              *checkpoint.CP
	interval        time.Duration
	offset          int64
	committedOffset int64
	committedAt     time.Time
	recordCount     int
}

// Close stops the worker process
func (p *publisher) Close() error {
	p.cancel()
	<-p.done
	return nil
}

// Check request a check from the publisher
func (p *publisher) Check() {
	select {
	case p.check <- struct{}{}:
	default:
	}

}

// Done allows external tools to signal off of when the publisher is done
func (p *publisher) Done() <-chan struct{} {
	return p.done
}

func (p *publisher) checkOnce() {
	segment, ctx := tracer.NewSegment(p.ctx, "publisher:check_once")
	defer segment.Finish()

	if p.offset == -1 {
		v, err := p.cp.Load(ctx, p.cpKey)
		if err != nil {
			return
		}
		p.offset = v
	}

	// read  events
	events, err := p.r.Read(ctx, p.offset+1, p.recordCount)
	if err != nil {
		return
	}

	// publish  events
	for _, event := range events {
		if err := p.h(event); err != nil {
			return
		}

		p.offset = event.Offset
	}

	// time to commit?
	if now := time.Now(); now.Sub(p.committedAt) > DefaultCommitInterval {
		if err := p.cp.Save(ctx, p.cpKey, p.offset); err == nil {
			p.committedAt = now
			p.committedOffset = p.offset
		}
	}
}

func (p *publisher) listenAndPublish() {
	defer close(p.done)
	defer close(p.check)
	defer p.segment.Finish()

	p.segment.Info("publisher:started", log.String("interval", p.interval.String()))

	timer := time.NewTicker(p.interval)
	defer timer.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-timer.C:
			p.checkOnce()
		case <-p.check:
			p.checkOnce()
		}
	}
}

// WithPublishEvents publishes received events to nats
func WithPublishEvents(fn func(eventsource.StreamRecord) error, nc *nats.Conn, env, boundedContext string) func(eventsource.StreamRecord) error {
	rootSubject := AggregateSubject(env, boundedContext) + "."

	return func(event eventsource.StreamRecord) error {
		if err := fn(event); err != nil {
			return err
		}

		subject := rootSubject + event.AggregateID
		go nc.Publish(subject, []byte(event.AggregateID))
		return nil
	}
}

// PublishStream reads from a stream and publishes
func PublishStream(ctx context.Context, h func(eventsource.StreamRecord) error, r eventsource.StreamReader, cp *checkpoint.CP, env, bc string) Publisher {
	cpKey := makeCheckpointKey(env, bc)
	segment, _ := tracer.NewSegment(ctx, "publish_stream", log.String("checkpoint", cpKey))

	child, cancel := context.WithCancel(ctx)
	p := &publisher{
		ctx:             child,
		cancel:          cancel,
		done:            make(chan struct{}),
		check:           make(chan struct{}, 1),
		segment:         segment,
		r:               r,
		h:               h,
		cpKey:           cpKey,
		cp:              cp,
		interval:        DefaultPublishInterval,
		offset:          -1,
		committedOffset: -1,
		recordCount:     100,
	}

	go p.listenAndPublish()

	return p
}

// WithReceiveNotifications listens to nats for notices on the AggregateSubject and prods the publisher
func WithReceiveNotifications(p Publisher, nc *nats.Conn, env, boundedContext string) Publisher {
	go func() {
		subject := NoticesSubject(env, boundedContext)
		fn := func(m *nats.Msg) {
			p.Check()
		}

		var sub *nats.Subscription
		for {
			select {
			case <-p.Done():
				return
			default:
			}

			if v, err := nc.Subscribe(subject, fn); err == nil {
				sub = v
				fmt.Println("ok")
				break
			}
		}

		<-p.Done()
		sub.Unsubscribe()
	}()

	return p
}

// PublishStan publishes events to the nats stream identified with the env and boundedContext
func PublishStan(st stan.Conn, subject string) func(event eventsource.StreamRecord) error {
	return func(event eventsource.StreamRecord) error {
		return st.Publish(subject, event.Data)
	}
}

// PublishStreamSingleton is similar to PublishStream except that there may be only one running in the environment
func PublishStreamSingleton(ctx context.Context, r eventsource.StreamReader, cp *checkpoint.CP, env, bc string, nc *nats.Conn) error {
	subject := AggregateSubject(env, bc)

	segment, ctx := tracer.NewSegment(ctx, "publish_stream")
	segment.SetBaggageItem("subject", subject)
	defer segment.Finish()

	st, err := stan.Connect(ClusterID, ksuid.New().String(), stan.NatsConn(nc))
	if err != nil {
		return errors.Wrapf(err, "unable to connec to nats streaming for subject, %v", subject)
	}

	t := ticker.Nats(nc, makeTickerSubject(env, bc))
	fn := action.Singleton(t, func(ctx context.Context) error {
		h := PublishStan(st, subject)
		h = WithPublishEvents(h, nc, env, bc) // publish events to here
		if env == "local" {
			h = withLogging(h, segment)
		}
		publisher := PublishStream(ctx, h, r, cp, env, bc)           // go!
		publisher = WithReceiveNotifications(publisher, nc, env, bc) // ping the publisher when events received
		<-publisher.Done()                                           // wait until done
		return nil
	})

	return fn(ctx)
}

func withLogging(fn func(event eventsource.StreamRecord) error, segment tracer.Segment) func(event eventsource.StreamRecord) error {
	return func(event eventsource.StreamRecord) error {
		segment.Info("publish_event:trace", log.String("id", event.AggregateID), log.Int("version", event.Version))
		return fn(event)
	}
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
