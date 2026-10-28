package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// relay turns the signals the ant command receives into the run's interrupts. While Os.run has
// a child in the foreground it holds back SIGINT and SIGQUIT, which the terminal sends the child
// too, and passes one on only when the child dies of it (stdlib.md, Os): a pager quit with Ctrl-C
// does not end the script, and a loop of make runs does end. Held signals do not queue, so one
// passed on counts once, however many arrived.
type relay struct {
	sigs       chan os.Signal
	quits      chan os.Signal // SIGQUIT, watched only while a child is held
	interrupts chan struct{}
	mu         sync.Mutex
	holding    int
	held       map[os.Signal]int // the signals held back so far, by kind
	passed     map[os.Signal]int // the count of each when one was last passed on
	last       os.Signal         // the signal behind the last interrupt sent
}

// newRelay starts the relay. A SIGINT the command was started with ignored, as nohup and a
// non-interactive shell's & start it, stays ignored, for the command and for its children.
func newRelay() *relay {
	r := &relay{
		sigs: make(chan os.Signal, 1), quits: make(chan os.Signal, 1), interrupts: make(chan struct{}, 1),
		held: map[os.Signal]int{}, passed: map[os.Signal]int{},
	}
	watched := []os.Signal{syscall.SIGTERM}
	if !signal.Ignored(os.Interrupt) {
		watched = append(watched, os.Interrupt)
	}
	signal.Notify(r.sigs, watched...)
	go func() {
		for {
			select {
			case sig, ok := <-r.sigs:
				if !ok {
					return
				}
				r.received(sig)
			case sig := <-r.quits:
				r.received(sig)
			}
		}
	}()
	return r
}

func (r *relay) stop() { signal.Stop(r.sigs) }

func (r *relay) received(sig os.Signal) {
	r.mu.Lock()
	if r.holding > 0 && (sig == os.Interrupt || sig == quit) {
		r.held[sig]++
		r.mu.Unlock()
		return
	}
	r.last = sig
	r.mu.Unlock()
	if sig == quit {
		dieBy(quit) // a SIGQUIT that slipped in as the hold ended does what it does without one
		return
	}
	r.send()
}

func (r *relay) send() {
	select {
	case r.interrupts <- struct{}{}:
	default:
	}
}

// hold is ant.WithForeground's: it holds back the keyboard's signals until the child exits, and
// then passes one on that arrived meanwhile when the child died of it: a SIGINT as the run's
// interrupt, which the run delivers itself, and a SIGQUIT by doing what it does unheld.
func (r *relay) hold() func(os.Signal) bool {
	r.mu.Lock()
	if r.holding == 0 && quit != nil && !signal.Ignored(quit) {
		signal.Notify(r.quits, quit)
	}
	r.holding++
	start := map[os.Signal]int{os.Interrupt: r.held[os.Interrupt], quit: r.held[quit]}
	r.mu.Unlock()
	return func(ended os.Signal) bool {
		r.mu.Lock()
		r.holding--
		if r.holding == 0 && quit != nil {
			signal.Stop(r.quits)
		}
		pass := ended != nil && r.held[ended] > start[ended] && r.held[ended] > r.passed[ended]
		if pass {
			r.passed[ended] = r.held[ended]
			r.last = ended
		}
		r.mu.Unlock()
		if pass && ended == quit {
			dieBy(quit)
		}
		return pass
	}
}

// ending is the signal the command ends by when an interrupt ended the run.
func (r *relay) ending() os.Signal {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.last
}
