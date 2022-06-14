package ddd

import (
	"context"
	"fmt"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/zhenyu888/ddd-core/apperr"
	"github.com/zhenyu888/ddd-core/diff"
	"github.com/zhenyu888/ddd-core/funcs"
)

type Repository interface {
	NextIdentify(context.Context) (int64, error)
	Save(context.Context, Aggregate) error
	Find(context.Context, int64) (Aggregate, error)
	FindNonNil(context.Context, int64) (Aggregate, error)
	Remove(context.Context, Aggregate) error
}

type RepositoryManager struct {
	pub   DomainEventPublisher
	idGen IdGenerator
}

var RepositoryManagerName = "ddd:core:RepositoryManager"

func NewRepositoryManager(pub DomainEventPublisher, idGen IdGenerator) *RepositoryManager {
	rlt := LoadOrStoreComponent(&RepositoryManager{}, func() interface{} {
		return &RepositoryManager{pub: pub, idGen: idGen}
	})
	return rlt.(*RepositoryManager)
}

func (r *RepositoryManager) Name() string {
	return RepositoryManagerName
}

func (r *RepositoryManager) NextIdentify(ctx context.Context) (int64, error) {
	return r.idGen.Gen(ctx)
}

func (r *RepositoryManager) AroundSave(ctx context.Context, agg Aggregate, doSave func(diff.AggregateDiff) error) error {
	r.AssertPointer(agg)
	if root, ok := agg.(AggregateRoot); ok {
		for _, event := range root.Events() {
			err := r.pub.Publish(ctx, event)
			if err != nil {
				return err
			}
		}
		root.ClearEvents()
		ad := root.Diff()
		err := doSave(ad)
		if err == nil {
			root.Attach(root)
		}
		return err
	}
	return doSave(diff.EmptyAggregateDiff())
}

func (r *RepositoryManager) AroundFind(ctx context.Context, doFind func() (Aggregate, error)) (Aggregate, error) {
	agg, err := doFind()
	if err == nil {
		if root, ok := agg.(AggregateRoot); ok {
			root.Attach(root)
		}
	}
	return agg, err
}

func (r *RepositoryManager) AroundRemove(ctx context.Context, agg Aggregate, doRemove func() error) error {
	r.AssertPointer(agg)
	err := doRemove()
	if err == nil {
		if root, ok := agg.(AggregateRoot); ok {
			root.Detach(root.AggregateId())
		}
	}
	return err
}

func (r *RepositoryManager) NonNil(agg Aggregate, err error) error {
	if err != nil {
		return err
	}
	if agg == nil || agg.AggregateId() <= 0 {
		notFound := fmt.Sprintf("%s not found", funcs.ReflectValueName(agg))
		return apperr.ErrNotFound(notFound, "", "")
	}
	return nil
}

func (r *RepositoryManager) AssertPointer(agg Aggregate) {
	if reflect.TypeOf(agg).Kind() != reflect.Ptr {
		panic("Aggregate param should be a pointer")
	}
}

func (r *RepositoryManager) AssertType(x, y interface{}) {
	if !funcs.TypeEqual(x, y) {
		panic(fmt.Sprintf("%s can not convert to %s", funcs.ReflectValueName(x), funcs.ReflectValueName(y)))
	}
}

type DBRepository interface {
	Repository
	GetDB(ctx context.Context) *gorm.DB
	GetWriteDB(ctx context.Context) *gorm.DB
}

// DBFactory 用来获取一个 *gorm.DB, 具体实现在edu/common_infra库里
type DBFactory interface {
	// LookupDB 自动选择主从，详细参考dbresolver
	LookupDB(context.Context) (*gorm.DB, error)
	// LookupWriteDB 强制获取写DB
	LookupWriteDB(context.Context) (*gorm.DB, error)
}

type AggregateExporter func() Aggregate

type DBRepositoryManager struct {
	*RepositoryManager
	dbFactory DBFactory
	exporter  AggregateExporter
}

func NewDBRepositoryManager(manager *RepositoryManager, factory DBFactory, exporter AggregateExporter) *DBRepositoryManager {
	return &DBRepositoryManager{
		RepositoryManager: manager,
		dbFactory:         factory,
		exporter:          exporter,
	}
}

func (r *DBRepositoryManager) GetDB(ctx context.Context) *gorm.DB {
	if txCtx, ok := ctx.(*TransactionContext); ok && txCtx.InTransaction() {
		return txCtx.TxDB()
	}
	db, _ := r.dbFactory.LookupDB(ctx)
	return db
}

func (r *DBRepositoryManager) GetWriteDB(ctx context.Context) *gorm.DB {
	if txCtx, ok := ctx.(*TransactionContext); ok && txCtx.InTransaction() {
		return txCtx.TxDB()
	}
	db, _ := r.dbFactory.LookupWriteDB(ctx)
	return db
}

func (r *DBRepositoryManager) Save(ctx context.Context, aggregate Aggregate) error {
	r.AssertType(aggregate, r.exporter())
	return r.AroundSave(ctx, aggregate, func(diff diff.AggregateDiff) error {
		db := r.GetDB(ctx)
		return db.Clauses(clause.OnConflict{UpdateAll: true}).Create(aggregate).Error
	})
}

func (r *DBRepositoryManager) Remove(ctx context.Context, aggregate Aggregate) error {
	r.AssertType(aggregate, r.exporter())
	return r.AroundRemove(ctx, aggregate, func() error {
		db := r.GetDB(ctx)
		return db.Delete(aggregate).Error
	})
}

func (r *DBRepositoryManager) Find(ctx context.Context, id int64) (Aggregate, error) {
	return r.AroundFind(ctx, func() (Aggregate, error) {
		db := r.GetDB(ctx)
		rlt := r.exporter()
		err := db.Limit(1).Find(rlt, id).Error
		if err != nil || rlt.AggregateId() <= 0 {
			return nil, err
		}
		return rlt, nil
	})
}

func (r *DBRepositoryManager) FindNonNil(ctx context.Context, id int64) (Aggregate, error) {
	rlt, err := r.Find(ctx, id)
	return rlt, r.NonNil(rlt, err)
}
