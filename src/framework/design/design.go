package design

import (
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"framework/design/tools/util"
	"framework/design/tools/http_api"
	"framework/design/tools/version"
	"fmt"
	"github.com/michaeljs1990/sqlitestore"
)

type DESIGN struct {
	sync.RWMutex

	opts atomic.Value

	startTime time.Time

	httpListener net.Listener

	exitChan chan int
	waitGroup util.WaitGroupWrapper

	store *sqlitestore.SqliteStore
}

func New(opts *Options) *DESIGN {
	d := &DESIGN{
		startTime: time.Now(),
		exitChan: make(chan int),
	}
	d.swapOpts(opts)

	d.logf(version.String("design"))

	return d
}


func (d *DESIGN) logf(f string, args ...interface{}) {
	if d.getOpts().Logger == nil {
		return
	}
	d.getOpts().Logger.Output(2, fmt.Sprintf(f, args...))
}

func (d *DESIGN) getOpts() *Options {
	return d.opts.Load().(*Options)
}

func (d *DESIGN) swapOpts(opts *Options) {
	d.opts.Store(opts)
}

func (d *DESIGN) RealHTTPAddr() *net.TCPAddr {
	d.RLock()
	defer d.RUnlock()
	return d.httpListener.Addr().(*net.TCPAddr)
}

func (d *DESIGN) GetStartTime() time.Time {
	return d.startTime
}

func (d *DESIGN) Main() {
	var store, err = sqlitestore.NewSqliteStore("./db/sessiondb", "sessions", "/", 3600, []byte("<SecretKey>"))
	if err != nil {
		d.logf("OPEN SQLITE DB FAILED")
		os.Exit(1)
	}

	d.Lock()
	d.store = store
	d.Unlock()

	var httpListener net.Listener

	ctx := &context{d}

	httpListener, err = net.Listen("tcp", d.getOpts().HTTPAddress)
	if err != nil {
		d.logf("FATAL: listen (%s) failed - %s", d.getOpts().HTTPAddress, err)
		os.Exit(1)
	}

	d.Lock()
	d.httpListener = httpListener
	d.Unlock()

	httpServer := newHTTPServer(ctx)

	d.waitGroup.Wrap(func() {
		http_api.Serve(httpListener, httpServer, "DesignServer", d.getOpts().Logger)
	})
}

func (d *DESIGN) Exit() {
	if d.httpListener != nil {
		d.httpListener.Close()
	}

	close(d.exitChan)
	d.waitGroup.Wait()
}

