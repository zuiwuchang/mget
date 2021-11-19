package widget

type Layout interface {
	Rect() (x, y, w, h int)
}
type layoutHandler struct {
	f func() (x, y, w, h int)
}

func (l layoutHandler) Rect() (x, y, w, h int) {
	return l.f()
}
func NewLayout(f func() (x, y, w, h int)) Layout {
	return layoutHandler{
		f: f,
	}
}
