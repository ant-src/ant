package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// lag is how long a signal may take to reach the relay after the kernel delivered it: the
// terminal's Ctrl-C reaches the command with the child, but on a loaded machine the relay's
// goroutine may see it only after the child has been reaped. A keyboard signal that late is
// still the child's.
const lag = 250 * time.Millisecond

// relay turns the signals the ant command receives into the run's interrupts. While Cmd.run has
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
	tail       time.Time         // until when a keyboard signal still counts as held, after a hold
	arrived    chan struct{}     // has a value once a signal is held
}

// newRelay starts the relay. A SIGINT the command was started with ignored, as nohup and a
// non-interactive shell's & start it, stays ignored, for the command and for its children.
func newRelay() *relay {
	r := &relay{
		sigs: make(chan os.Signal, 1), quits: make(chan os.Signal, 1), interrupts: make(chan struct{}, 1),
		held: map[os.Signal]int{}, passed: map[os.Signal]int{}, arrived: make(chan struct{}, 1),
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
	if (r.holding > 0 || time.Now().Before(r.tail)) && (sig == os.Interrupt || sig == quit) {
		r.held[sig]++
		r.mu.Unlock()
		select {
		case r.arrived <- struct{}{}:
		default:
		}
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
// interrupt, which the run delivers itself, and a SIGQUIT by doing what it does unheld. A child
// that died of one the relay has not seen yet waits up to lag for it, and one that arrives within
// lag after the hold ends is held too.
func (r *relay) hold() func(os.Signal) bool {
	r.mu.Lock()
	if r.holding == 0 && quit != nil && !signal.Ignored(quit) {
		signal.Notify(r.quits, quit)
	}
	r.holding++
	start := map[os.Signal]int{os.Interrupt: r.held[os.Interrupt], quit: r.held[quit]}
	r.mu.Unlock()
	return func(ended os.Signal) bool {
		if ended == os.Interrupt || ended == quit {
			r.await(ended, start[ended])
		}
		r.mu.Lock()
		r.holding--
		if r.holding == 0 {
			r.tail = time.Now().Add(lag)
			if quit != nil {
				time.AfterFunc(lag, r.stopQuits)
			}
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

// await waits up to lag for the relay to hold a sig beyond count.
func (r *relay) await(sig os.Signal, count int) {
	deadline := time.NewTimer(lag)
	defer deadline.Stop()
	for {
		r.mu.Lock()
		seen := r.held[sig] > count
		r.mu.Unlock()
		if seen {
			return
		}
		select {
		case <-r.arrived:
		case <-deadline.C:
			return
		}
	}
}

// stopQuits stops watching SIGQUIT once no child is held and the lag after the last has passed.
func (r *relay) stopQuits() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.holding == 0 && !time.Now().Before(r.tail) {
		signal.Stop(r.quits)
	}
}

// ending is the signal the command ends by when an interrupt ended the run.
func (r *relay) ending() os.Signal {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.last
}
