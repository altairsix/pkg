package eventsourcex

import (
	"context"
	"fmt"
	"time"

	"github.com/altairsix/eventsource"
	"github.com/altairsix/pkg/checkpoint"
	nats "github.com/nats-io/go-nats"
)

const (
	// DefaultCommitInterval the minimum amount of time that must pass between offset commits
	DefaultCommitInterval = time.Second * 3
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
	if p.offset == -1 {
		v, err := p.cp.Load(p.ctx, p.cpKey)
		if err != nil {
			return
		}
		p.offset = v
	}

	// read  events
	events, err := p.r.Read(p.ctx, p.offset, p.recordCount)
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
		if err := p.cp.Save(p.ctx, p.cpKey, p.offset); err == nil {
			p.committedAt = now
			p.committedOffset = p.offset
		}
	}
}

func (p *publisher) listenAndPublish() {
	defer close(p.done)
	defer close(p.check)

	timer := time.NewTimer(p.interval)
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

// PublishEvents publishes received events to nats
func PublishEvents(nc *nats.Conn, env, boundedContext string) func(eventsource.StreamRecord) error {
	rootSubject := AggregateSubject(env, boundedContext) + "."

	return func(event eventsource.StreamRecord) error {
		subject := rootSubject + event.AggregateID
		go nc.Publish(subject, nil)
		return nil
	}
}

// PublishStream reads from a stream and publishes
func PublishStream(ctx context.Context, h func(eventsource.StreamRecord) error, r eventsource.StreamReader, cp *checkpoint.CP, cpKey string) Publisher {
	child, cancel := context.WithCancel(ctx)
	p := &publisher{
		ctx:             child,
		cancel:          cancel,
		done:            make(chan struct{}),
		check:           make(chan struct{}, 1),
		r:               r,
		h:               h,
		cpKey:           cpKey,
		cp:              cp,
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
		subject := AggregateSubject(env, boundedContext)
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
