package diff

type AggregateDiff interface {
	IsEmpty() bool
	IsSelfChanged() bool
	GetDiff(string) Diff
	GetListDiff(string) ListDiff
}

type AggregateDiffBuilder interface {
	AggregateDiff
	SetSelfChanged(bool) AggregateDiffBuilder
	PutDiff(string, Diff) AggregateDiffBuilder
	PutListDiff(string, ListDiff) AggregateDiffBuilder
	Build() AggregateDiff
}

type Diff interface {
	IsChanged() bool
}

type ListDiff interface {
	Diff
	Added() []interface{}
	Removed() []interface{}
	Modified() []interface{}
}

type ListDiffBuilder interface {
	AppendAdded(v interface{}) ListDiffBuilder
	AppendRemoved(v interface{}) ListDiffBuilder
	AppendModified(v interface{}) ListDiffBuilder
	Build() ListDiff
}
