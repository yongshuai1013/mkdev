package tabs

import (
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/venkatkrishna07/mkdev/internal/store"
	"github.com/venkatkrishna07/mkdev/internal/tui/msg"
	"github.com/venkatkrishna07/mkdev/internal/tui/styles"
)

// RTTSource resolves the rolling RTT window for a domain. nil = no live RTT.
type RTTSource func(domain string) []time.Duration

type Domains struct {
	th     styles.Theme
	width  int
	height int
	table  table.Model
	routes []store.Route
	rtt    RTTSource
}

func NewDomains(th styles.Theme, width, height int) Domains {
	return NewDomainsWithRTT(th, width, height, nil)
}

func NewDomainsWithRTT(th styles.Theme, width, height int, rtt RTTSource) Domains {
	d := Domains{th: th, width: width, height: height, rtt: rtt}
	t := table.New(
		table.WithColumns(d.layoutCols(d.tableWidth())),
		table.WithFocused(true),
		table.WithHeight(8),
	)
	t.SetStyles(tableStyles(th))
	d.table = t
	return d
}

func tableStyles(th styles.Theme) table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		Bold(true).
		Foreground(th.Primary).
		BorderStyle(lipgloss.HiddenBorder()).
		BorderBottom(false).
		Padding(0, 1)
	s.Selected = s.Selected.
		Bold(true).
		Foreground(th.OnPill).
		Background(th.Accent)
	s.Cell = s.Cell.Padding(0, 1)
	return s
}

const (
	domainsStatusW = 12
	domainsSourceW = 12
	domainsAddedW  = 12
)

func (d *Domains) tableWidth() int {
	w := d.width
	if w <= 0 {
		w = 100
	}
	return w
}

func (d Domains) layoutCols(tableW int) []table.Column {
	fixed := domainsStatusW + domainsSourceW + domainsAddedW
	rem := tableW - fixed - 10 // padding budget across cols
	if rem < 20 {
		rem = 20
	}
	domW := rem / 2
	tgtW := rem - domW
	return []table.Column{
		{Title: "DOMAIN", Width: domW},
		{Title: "TARGET", Width: tgtW},
		{Title: "STATUS", Width: domainsStatusW},
		{Title: "SOURCE", Width: domainsSourceW},
		{Title: "ADDED", Width: domainsAddedW},
	}
}

func (d Domains) Title() string { return "Domains" }

func (d Domains) Init() tea.Cmd { return nil }

func (d Domains) Update(in tea.Msg) (Domains, tea.Cmd) {
	switch m := in.(type) {
	case msg.RoutesRefreshed:
		d.routes = m.Routes
		d.refreshRows()
	case tea.WindowSizeMsg:
		d.width = m.Width
		d.height = m.Height
		d.table.SetColumns(d.layoutCols(d.tableWidth()))
		d.fitHeight()
	}
	var cmd tea.Cmd
	d.table, cmd = d.table.Update(in)
	return d, cmd
}

func (d *Domains) fitHeight() {
	rows := max(len(d.routes), 1)
	budget := max(d.height-6, 3)
	h := min(rows+1, budget)
	d.table.SetHeight(h)
}

func (d *Domains) refreshRows() {
	if len(d.routes) == 0 {
		d.table.SetRows(nil)
		return
	}
	rows := make([]table.Row, len(d.routes))
	for i, r := range d.routes {
		rows[i] = table.Row{
			r.Domain,
			r.Target,
			d.statusCell(r),
			r.Source,
			r.AddedAt.Format("2006-01-02"),
		}
	}
	d.table.SetRows(rows)
	d.fitHeight()
}

func (d Domains) statusCell(r store.Route) string {
	var pill string
	switch {
	case !r.Enabled:
		pill = d.th.PillOff.Render("● off")
	case d.rtt != nil && len(d.rtt(r.Domain)) > 0:
		pill = d.th.PillUp.Render("● live")
	default:
		pill = d.th.PillUp.Render("● up")
	}
	if r.Shared {
		return pill + " " + d.th.Dim.Render("LAN")
	}
	return pill
}

func (d Domains) View() string {
	if len(d.routes) == 0 {
		hint := d.th.Dim.Render("no routes yet — press ") + d.th.FooterKey.Render("a") + d.th.Dim.Render(" to add")
		return lipgloss.JoinVertical(lipgloss.Left, hint, d.table.View())
	}
	return d.table.View()
}

func (d Domains) Selected() (store.Route, bool) {
	if len(d.routes) == 0 {
		return store.Route{}, false
	}
	idx := min(max(d.table.Cursor(), 0), len(d.routes)-1)
	return d.routes[idx], true
}
