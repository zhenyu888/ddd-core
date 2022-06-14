package ddd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TransactionPropagation int8

const (
	PropagationRequired    = iota // 支持当前事务，如果当前没有事务，就新建一个事务
	PropagationRequiresNew        // 新建事务，如果当前存在事务，把当前事务挂起，两个事务互不影响
	PropagationNested             // 支持当前事务，如果当前事务存在，则执行一个嵌套事务，如果当前没有事务，就新建一个事务
	PropagationNever              // 以非事务方式执行，如果当前存在事务，直接返回错误
)

func isPropagationSupport(propagation TransactionPropagation) bool {
	switch propagation {
	case PropagationRequired:
		return true
	case PropagationRequiresNew:
		return true
	case PropagationNested:
		return true
	case PropagationNever:
		return true
	}
	return false
}

func defaultPropagation() TransactionPropagation {
	return PropagationRequired
}

var (
	ErrNotInTransaction = errors.New("not in transaction, can't commit")
	ErrInTransaction    = errors.New("never propagation should not in transaction")
)

type TransactionContext struct {
	ctx    context.Context
	tx     *gorm.DB
	parent *TransactionContext
}

func (c *TransactionContext) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *TransactionContext) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *TransactionContext) Err() error {
	return c.ctx.Err()
}

func (c *TransactionContext) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func (c *TransactionContext) IsRoot() bool {
	return c.parent == nil
}

func (c *TransactionContext) Ctx() context.Context {
	return c.ctx
}

func (c *TransactionContext) TxDB() *gorm.DB {
	return c.tx
}

func (c *TransactionContext) TxError() error {
	if c.tx != nil {
		return c.tx.Error
	}
	return nil
}

func (c *TransactionContext) InTransaction() bool {
	if c.tx == nil {
		return false
	}
	committer, ok := c.tx.Statement.ConnPool.(gorm.TxCommitter)
	return ok && committer != nil
}

func (c *TransactionContext) Session(config *gorm.Session) *TransactionContext {
	return &TransactionContext{
		ctx:    c.ctx,
		tx:     c.tx.Session(config),
		parent: c,
	}
}

func (c *TransactionContext) Rollback() {
	if c.InTransaction() {
		c.tx.Rollback()
	}
}

func (c *TransactionContext) Commit() error {
	if !c.InTransaction() {
		return ErrNotInTransaction
	}
	if c.IsRoot() {
		return c.tx.Commit().Error
	}
	return nil
}

type TransactionManager struct {
	factory DBFactory
}

var TransactionManagerName = "ddd:core:TransactionManager"

func NewTransactionManager(factory DBFactory) *TransactionManager {
	rlt := LoadOrStoreComponent(&TransactionManager{}, func() interface{} {
		return &TransactionManager{factory: factory}
	})
	return rlt.(*TransactionManager)
}

func (m *TransactionManager) Name() string {
	return TransactionManagerName
}

func (m *TransactionManager) lookupDB(ctx context.Context) (*gorm.DB, error) {
	if txCtx, ok := ctx.(*TransactionContext); ok && txCtx.tx != nil {
		return txCtx.tx, nil
	}
	return m.lookupNewDB(ctx)
}

func (m *TransactionManager) lookupNewDB(ctx context.Context) (*gorm.DB, error) {
	return m.factory.LookupDB(ctx)
}

func (m *TransactionManager) Transaction(ctx context.Context, bizFn func(txCtx context.Context) error, propagations ...TransactionPropagation) error {
	propagation := defaultPropagation()
	if len(propagations) > 0 && isPropagationSupport(propagations[0]) {
		propagation = propagations[0]
	}
	switch propagation {
	case PropagationNever:
		return m.withNeverPropagation(ctx, bizFn)
	case PropagationNested:
		return m.withNestedPropagation(ctx, bizFn)
	case PropagationRequired:
		return m.withRequiredPropagation(ctx, bizFn)
	case PropagationRequiresNew:
		return m.withRequiresNewPropagation(ctx, bizFn)
	}
	panic("not support propagation")
}

func (m *TransactionManager) withNeverPropagation(ctx context.Context, bizFn func(txCtx context.Context) error) error {
	if txCtx, ok := ctx.(*TransactionContext); ok && txCtx.InTransaction() {
		return ErrInTransaction
	}
	return bizFn(ctx)
}

func (m *TransactionManager) withNestedPropagation(ctx context.Context, bizFn func(txCtx context.Context) error) error {
	var err error
	if txCtx, ok := ctx.(*TransactionContext); ok && txCtx.InTransaction() {
		panicked := true
		db := txCtx.TxDB()
		if !db.DisableNestedTransaction {
			err = db.SavePoint(fmt.Sprintf("sp%p", bizFn)).Error
			defer func() {
				// Make sure to rollback when panic, Block error or Commit error
				if panicked || err != nil {
					db.RollbackTo(fmt.Sprintf("sp%p", bizFn))
				}
			}()
		}
		if err == nil {
			err = bizFn(txCtx.Session(&gorm.Session{}))
		}
		panicked = false
	} else {
		err = m.withRequiredPropagation(ctx, bizFn)
	}
	return err
}

func (m *TransactionManager) withRequiredPropagation(ctx context.Context, bizFn func(txCtx context.Context) error) error {
	var err error
	panicked := true
	if txCtx, ok := ctx.(*TransactionContext); ok && txCtx.InTransaction() {
		defer func() {
			if panicked || err != nil {
				txCtx.Rollback()
			}
		}()
		err = bizFn(txCtx.Session(&gorm.Session{}))
	} else {
		var db *gorm.DB
		db, err = m.lookupDB(ctx)
		if err != nil {
			return err
		}
		if !ok {
			txCtx = &TransactionContext{
				ctx: ctx,
				tx:  db.Begin(),
			}
		} else {
			txCtx.tx = db.Begin()
		}
		defer func() {
			if panicked || err != nil {
				txCtx.Rollback()
			}
		}()
		if err = txCtx.TxError(); err == nil {
			err = bizFn(txCtx)
		}

		if err == nil {
			err = txCtx.Commit()
		}
	}
	panicked = false
	return err
}

func (m *TransactionManager) withRequiresNewPropagation(ctx context.Context, bizFn func(txCtx context.Context) error) error {
	panicked := true
	var pureCtx context.Context
	if txCtx, ok := ctx.(*TransactionContext); ok {
		pureCtx = txCtx.Ctx()
	} else {
		pureCtx = ctx
	}
	db, err := m.lookupNewDB(pureCtx)
	if err != nil {
		return err
	}
	txCtx := &TransactionContext{
		ctx: ctx,
		tx:  db.Begin(),
	}
	defer func() {
		if panicked || err != nil {
			txCtx.Rollback()
		}
	}()
	if err = txCtx.TxError(); err == nil {
		err = bizFn(txCtx)
	}

	if err == nil {
		err = txCtx.Commit()
	}
	panicked = false
	return err
}
