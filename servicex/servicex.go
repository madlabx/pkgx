package servicex

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/madlabx/pkgx/dbc"
)

type CommonSrv struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Dbc    *dbc.DbClient
}

func (cs *CommonSrv) Init(pCtx context.Context, dbc *dbc.DbClient) {
	cs.Ctx, cs.Cancel = context.WithCancel(pCtx)
	cs.Dbc = dbc
}

type ServiceIf interface {
	Launch(ctx context.Context, sql *dbc.DbClient) error
	GetName() string
}

type serviceX struct {
	srvs  map[string]ServiceIf
	pCtx  context.Context
	sql   *dbc.DbClient
	cache any
}

type ServiceErrorHandleIf interface {
	ErrorHandle() (bool, error)
}

var (
	once sync.Once
	sx   serviceX
)

func Launch(pctx context.Context, sql *dbc.DbClient, ss ...ServiceIf) error {
	var err error
	once.Do(func() {
		sx.pCtx = pctx
		sx.sql = sql

		sx.srvs = make(map[string]ServiceIf, len(ss))
		for _, srv := range ss {
			if reflect.TypeOf(srv).Kind() != reflect.Pointer {
				err = fmt.Errorf("Invalid service type, srv:%v", srv)
				return
			}

			err = srv.Launch(sx.pCtx, sx.sql)
			if err != nil {
				eh, ok := srv.(ServiceErrorHandleIf)
				if ok {
					var toBreak bool
					if toBreak, err = eh.ErrorHandle(); toBreak {
						return
					}
				} else {
					return
				}
			}

		}
		for _, s := range ss {
			sx.srvs[s.GetName()] = s
		}

		err = run()
	})

	return err
}

func Get(ss ServiceIf) ServiceIf {
	return sx.srvs[ss.GetName()]
}

func run() (err error) {

	return nil
}
