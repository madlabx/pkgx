package graceful

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/madlabx/pkgx/log"
)

type GracefulService interface {
	Stop() error
	Name() string
}

type Graceful struct {
	ctx    context.Context
	cancel context.CancelFunc
	sigCh  chan os.Signal
}

func New(parent context.Context) *Graceful {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)

	gc := &Graceful{
		ctx:    ctx,
		cancel: cancel,
		sigCh:  make(chan os.Signal, 1),
	}

	go gc.listenToSignal()
	signal.Notify(gc.sigCh)

	return gc
}

func (gc *Graceful) Context() context.Context {
	return gc.ctx
}

func (gc *Graceful) listenToSignal() {

	defer func() {
		log.Errorf("listenToSignal defer")
		gc.cancel()
	}()
	log.Errorf("listenToSignal start")
	for {
		sig := <-gc.sigCh
		switch sig {

		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL:
			log.Errorf("receive signal %v, program will exit", sig.String())
			return
		case syscall.SIGSEGV:
			log.Errorf("receive signal %v, callstack: %v", sig.String(), debug.Stack())
			return
		default:
		}
	}
}

func (gc *Graceful) WaitToQuit(gss ...GracefulService) {
	log.Errorf("WaitToQuit")
	<-gc.ctx.Done()
	log.Errorf("WaitToQuit")

	quitCtx, quitCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)

	go func() {
		log.Infof("Start to stop services")
		var wg sync.WaitGroup
		for _, gs := range gss {
			wg.Add(1)
			go func(s GracefulService) {
				log.IgnoreErrf(s.Stop(), s.Name())
				wg.Done()
			}(gs)
		}
		wg.Wait()
		quitCtxCancel()
	}()

	<-quitCtx.Done()
}
