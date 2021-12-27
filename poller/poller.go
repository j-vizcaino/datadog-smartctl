package poller

import (
	"context"
	"sync"
	"time"

	"github.com/j-vizcaino/datadog-smartctl/smartctl"
)

type QueryDeviceFunc func(ctx context.Context, dev string) (smartctl.Data, error)
type OnNewDataFunc func(ctx context.Context, data smartctl.Data)

type Poller struct {
	pollingInterval time.Duration
	queryDevice     QueryDeviceFunc
	onNewData       OnNewDataFunc
	device          string
	stopChan        chan bool
	stopOnce        sync.Once
	running         sync.WaitGroup
}

func New(queryDevice QueryDeviceFunc, onNewData OnNewDataFunc, device string) *Poller {
	return &Poller{
		queryDevice: queryDevice,
		onNewData:   onNewData,
		device:      device,
		stopChan:    make(chan bool),
	}
}

func (p *Poller) Poll(ctx context.Context, pollingInterval time.Duration) {
	p.running.Add(1)
	go p.periodicPoll(ctx, pollingInterval)
}

func (p *Poller) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopChan)
	})
	p.running.Wait()
}

func (p *Poller) periodicPoll(ctx context.Context, pollingInterval time.Duration) {
	run := true
	p.pollAndReport(ctx)
	for run {
		select {
		case <-time.After(pollingInterval):
			p.pollAndReport(ctx)
		case <-ctx.Done():
			run = false
		case <-p.stopChan:
			run = false
		}
	}
	p.running.Done()
}

func (p *Poller) pollAndReport(ctx context.Context) {
	data, err := p.queryDevice(ctx, p.device)
	if err != nil {
		return
	}

	p.onNewData(ctx, data)
}
