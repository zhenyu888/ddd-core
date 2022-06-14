package diff

func NewDiff(isDiff bool) Diff {
	return diff(isDiff)
}

type diff bool

func (d diff) IsChanged() bool {
	return bool(d)
}

type listDiff struct {
	added    []interface{}
	removed  []interface{}
	modified []interface{}
}

func (l *listDiff) IsChanged() bool {
	return len(l.added) > 0 || len(l.removed) > 0 || len(l.modified) > 0
}

func (l *listDiff) Added() []interface{} {
	return l.added
}

func (l *listDiff) Removed() []interface{} {
	return l.removed
}

func (l *listDiff) Modified() []interface{} {
	return l.modified
}

func NewListDiffBuilder() ListDiffBuilder {
	return &listDiffBuilder{
		ld: &listDiff{},
	}
}

type listDiffBuilder struct {
	ld *listDiff
}

func (b *listDiffBuilder) AppendAdded(v interface{}) ListDiffBuilder {
	b.ld.added = append(b.ld.added, v)
	return b
}

func (b *listDiffBuilder) AppendRemoved(v interface{}) ListDiffBuilder {
	b.ld.removed = append(b.ld.removed, v)
	return b
}

func (b *listDiffBuilder) AppendModified(v interface{}) ListDiffBuilder {
	b.ld.modified = append(b.ld.modified, v)
	return b
}

func (b *listDiffBuilder) Build() ListDiff {
	return b.ld
}

type aggregateDiff struct {
	selfChanged bool
	diffMap     map[string]Diff
	listDiffMap map[string]ListDiff
}

func (d *aggregateDiff) IsEmpty() bool {
	if d.selfChanged {
		return false
	}
	for _, v := range d.diffMap {
		if v.IsChanged() {
			return false
		}
	}
	for _, v := range d.listDiffMap {
		if v.IsChanged() {
			return false
		}
	}
	return true
}

func (d *aggregateDiff) IsSelfChanged() bool {
	return d.selfChanged
}

func (d *aggregateDiff) GetDiff(name string) Diff {
	if rlt, ok := d.diffMap[name]; ok {
		return rlt
	}
	return NewDiff(false)
}

func (d *aggregateDiff) GetListDiff(name string) ListDiff {
	if rlt, ok := d.listDiffMap[name]; ok {
		return rlt
	}
	return EmptyListDiff()
}

func NewAggregateDiffBuilder() AggregateDiffBuilder {
	return &aggregateDiffBuilder{
		&aggregateDiff{
			diffMap:     make(map[string]Diff),
			listDiffMap: make(map[string]ListDiff),
		},
	}
}

type aggregateDiffBuilder struct {
	ad *aggregateDiff
}

func (b *aggregateDiffBuilder) IsEmpty() bool {
	return b.ad.IsEmpty()
}

func (b *aggregateDiffBuilder) IsSelfChanged() bool {
	return b.ad.IsSelfChanged()
}

func (b *aggregateDiffBuilder) GetDiff(tag string) Diff {
	return b.ad.GetDiff(tag)
}

func (b *aggregateDiffBuilder) GetListDiff(tag string) ListDiff {
	return b.ad.GetListDiff(tag)
}

func (b *aggregateDiffBuilder) SetSelfChanged(selfChange bool) AggregateDiffBuilder {
	b.ad.selfChanged = selfChange
	return b
}

func (b *aggregateDiffBuilder) PutDiff(tag string, d Diff) AggregateDiffBuilder {
	b.ad.diffMap[tag] = d
	return b
}

func (b *aggregateDiffBuilder) PutListDiff(tag string, ld ListDiff) AggregateDiffBuilder {
	b.ad.listDiffMap[tag] = ld
	return b
}

func (b *aggregateDiffBuilder) Build() AggregateDiff {
	return b.ad
}
