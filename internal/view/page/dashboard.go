package page

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/slok/grafterm/internal/controller"
	"github.com/slok/grafterm/internal/model"
	"github.com/slok/grafterm/internal/service/log"
	"github.com/slok/grafterm/internal/view/grid"
	"github.com/slok/grafterm/internal/view/page/widget"
	"github.com/slok/grafterm/internal/view/render"
	viewsync "github.com/slok/grafterm/internal/view/sync"
	"github.com/slok/grafterm/internal/view/template"
	"github.com/slok/grafterm/internal/view/variable"
)

// DashboardCfg is the configuration required to create a Dashboard.
type DashboardCfg struct {
	AppRelativeTimeRange time.Duration
	AppOverrideVariables map[string]string
	Controller           controller.Controller
	Dashboard            model.Dashboard
	Renderer             render.Renderer
}

// NewDashboard returns a new syncer from a dashboard with all the required
// widgets loaded.
// The widgets the dashboard manages at the same time are syncers also.
func NewDashboard(ctx context.Context, cfg DashboardCfg, logger log.Logger) (viewsync.Syncer, error) {
	// Create variablers.
	vs, err := variable.NewVariablers(variable.FactoryConfig{
		TimeRange: cfg.AppRelativeTimeRange,
		Dashboard: cfg.Dashboard,
	})

	// Create Grid.
	var gr *grid.Grid
	if cfg.Dashboard.Grid.FixedWidgets {
		gr, err = grid.NewFixedGrid(cfg.Dashboard.Grid.MaxWidth, cfg.Dashboard.Widgets)
		if err != nil {
			return nil, err
		}
	} else {
		gr, err = grid.NewAdaptiveGrid(cfg.Dashboard.Grid.MaxWidth, cfg.Dashboard.Widgets)
		if err != nil {
			return nil, err
		}
	}

	d := &dashboard{
		cfg:        cfg,
		variablers: vs,
		ctrl:       cfg.Controller,
		logger:     logger,
	}

	// Call the View to load the dashboard and return us the widgets that we will need to call.
	renderWidgets, err := cfg.Renderer.LoadDashboard(ctx, gr)
	if err != nil {
		return nil, err
	}

	d.widgets = d.createWidgets(renderWidgets)

	return d, nil
}

type dashboard struct {
	cfg        DashboardCfg
	widgets    []viewsync.Syncer
	ctrl       controller.Controller
	variablers map[string]variable.Variabler
	logger     log.Logger
}

func (d *dashboard) Sync(ctx context.Context, r *viewsync.Request) error {
	// Add dashboard sync data.
	r = d.syncData(r)

	// Create a context with timeout for widget sync operations
	widgetCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Sync all widgets with proper error handling and timeout
	var wg sync.WaitGroup
	errorChan := make(chan error, len(d.widgets))
	
	for _, w := range d.widgets {
		wg.Add(1)
		go func(widget viewsync.Syncer) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("widget sync panic recovered: %v", r)
				}
			}()

			// Don't wait to sync all at the same time, the widgets
			// should control multiple calls to sync and reject the sync
			// if already syncing.
			if widget == nil {
				errorChan <- fmt.Errorf("widget is nil, skipping sync")
				return
			}

			err := widget.Sync(widgetCtx, r)
			if err != nil {
				if widgetCtx.Err() == context.DeadlineExceeded {
					errorChan <- fmt.Errorf("widget sync timeout: %w", err)
					return
				}
				if widgetCtx.Err() == context.Canceled {
					errorChan <- fmt.Errorf("widget sync canceled: %w", err)
					return
				}
				errorChan <- fmt.Errorf("error syncing widget: %w", err)
			}
		}(w)
	}

	// Wait for all widgets to finish
	wg.Wait()
	close(errorChan)

	// Log any errors that occurred
	for err := range errorChan {
		d.logger.Errorf(err.Error())
	}

	return nil
}

func (d *dashboard) createWidgets(rws []render.Widget) []viewsync.Syncer {
	widgets := []viewsync.Syncer{}

	// Create app widgets based on the render view widgets.
	for _, rw := range rws {
		var w viewsync.Syncer

		// Depending on the type create a widget kind or another.
		switch v := rw.(type) {
		case render.GaugeWidget:
			w = widget.NewGauge(d.ctrl, v)
		case render.SinglestatWidget:
			w = widget.NewSinglestat(d.ctrl, v)
		case render.GraphWidget:
			w = widget.NewGraph(d.ctrl, v, d.logger)
		default:
			continue
		}

		// Dashboard data.
		dashboardData := d.staticData()
		overrideData := d.overrideVariableData()

		// Widget middlewares.
		w = withWidgetDataMiddleware(dashboardData, overrideData, w) // Assign static data to widget.

		widgets = append(widgets, w)
	}

	return widgets
}

func (d *dashboard) overrideVariableData() template.Data {
	od := map[string]interface{}{}
	for k, v := range d.cfg.AppOverrideVariables {
		od[k] = v
	}
	return template.Data(od)
}

func (d *dashboard) staticData() template.Data {
	// Load variablers data from the dashboard scope.
	dashboardData := map[string]interface{}{}
	for vid, v := range d.variablers {
		if v.Scope() == variable.ScopeDashboard {
			dashboardData[vid] = v.GetValue()
		}
	}

	return dashboardData
}

func (d *dashboard) syncData(r *viewsync.Request) *viewsync.Request {
	// Load variablers data from the sync scope.
	data := map[string]interface{}{}
	for vid, v := range d.variablers {
		if v.Scope() == variable.ScopeSync {
			data[vid] = v.GetValue()
		}
	}
	r.TemplateData = r.TemplateData.WithData(data)

	return r
}
