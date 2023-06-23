package fault

import "context"

type Orchestrator struct {
	agents    []Agent
	outputChs []chan Claim
	responses chan Response
}

func NewOrchestrator(maxDepth int, traces []TraceProvider, root Claim) Orchestrator {
	o := Orchestrator{
		responses: make(chan Response, 100),
		outputChs: make([]chan Claim, len(traces)),
		agents:    make([]Agent, len(traces)),
	}
	for i, trace := range traces {
		var game Game // TODO: Properly create this game
		o.agents[i] = NewAgent(game, maxDepth, trace, &o)
		o.outputChs[i] = make(chan Claim)
	}
	return o
}

func (o *Orchestrator) Respond(_ context.Context, response Response) error {
	o.responses <- response
	return nil
}

func (o *Orchestrator) Start() {
	// TODO handle shutdown
	go o.reponderThread()
	for i := 0; i < len(o.agents); i++ {
		go runAgent(&o.agents[i], o.outputChs[i])
	}
}

func runAgent(agent *Agent, claimCh <-chan Claim) {
	for {
		// TODO: Multiple claims / how to balance performing actions with
		// accepting new claims
		claim := <-claimCh
		agent.AddClaim(claim)
		agent.PerformActions()
	}
}

func (o *Orchestrator) reponderThread() {
	for {
		resp := <-o.responses
		for _, ch := range o.outputChs {
			ch <- resp.Parent // TODO: new claim
		}
	}
}
