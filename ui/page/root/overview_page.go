package root

import (
	"context"

	"gioui.org/layout"

	"github.com/crypto-power/cryptopower/app"
	"github.com/crypto-power/cryptopower/ui/cryptomaterial"
	"github.com/crypto-power/cryptopower/ui/load"
	"github.com/crypto-power/cryptopower/ui/values"
)

const (
	OverviewPageID = "Overview"
)

type OverviewPage struct {
	*app.GenericPageModal
	*load.Load

	ctx       context.Context
	ctxCancel context.CancelFunc
}

func NewOverviewPage(l *load.Load) *OverviewPage {
	op := &OverviewPage{
		Load:             l,
		GenericPageModal: app.NewGenericPageModal(OverviewPageID),
	}

	return op
}

// ID is a unique string that identifies the page and may be used
// to differentiate this page from other pages.
// Part of the load.Page interface.
func (op *OverviewPage) ID() string {
	return OverviewPageID
}

// OnNavigatedTo is called when the page is about to be displayed and
// may be used to initialize page features that are only relevant when
// the page is displayed.
// Part of the load.Page interface.
func (op *OverviewPage) OnNavigatedTo() {
	op.ctx, op.ctxCancel = context.WithCancel(context.TODO())
}

// HandleUserInteractions is called just before Layout() to determine
// if any user interaction recently occurred on the page and may be
// used to update the page's UI components shortly before they are
// displayed.
// Part of the load.Page interface.
func (op *OverviewPage) HandleUserInteractions() {

}

// OnNavigatedFrom is called when the page is about to be removed from
// the displayed window. This method should ideally be used to disable
// features that are irrelevant when the page is NOT displayed.
// NOTE: The page may be re-displayed on the app's window, in which case
// OnNavigatedTo() will be called again. This method should not destroy UI
// components unless they'll be recreated in the OnNavigatedTo() method.
// Part of the load.Page interface.
func (op *OverviewPage) OnNavigatedFrom() {
	op.ctxCancel()
}

// Layout draws the page UI components into the provided layout context
// to be eventually drawn on screen.
// Part of the load.Page interface.
func (op *OverviewPage) Layout(gtx C) D {
	op.Load.SetCurrentAppWidth(gtx.Constraints.Max.X)
	if op.Load.GetCurrentAppWidth() <= gtx.Dp(values.StartMobileView) {
		return op.layoutMobile(gtx)
	}
	return op.layoutDesktop(gtx)
}

func (op *OverviewPage) layoutDesktop(gtx C) D {
	return op.comingSoonLayout(gtx)
}

func (op *OverviewPage) layoutMobile(gtx C) D {
	return op.comingSoonLayout(gtx)
}

func (op *OverviewPage) comingSoonLayout(gtx C) D {
	return cryptomaterial.LinearLayout{
		Width:       cryptomaterial.MatchParent,
		Height:      cryptomaterial.MatchParent,
		Orientation: layout.Vertical,
		Alignment:   layout.Middle,
		Direction:   layout.Center,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			lblText := op.Theme.H4("Overview Page Coming soon.....")
			lblText.Color = op.Theme.Color.PageNavText
			return lblText.Layout(gtx)
		}),
	)
}
