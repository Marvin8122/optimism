package fault

import (
	"context"
	"sync"
)

type Agent struct {
	mu        sync.Mutex
	game      Game
	solver    *Solver
	trace     TraceProvider
	responder Responder
	maxDepth  int
}

func NewAgent(game Game, maxDepth int, trace TraceProvider, responder Responder) Agent {
	return Agent{
		game:      game,
		solver:    NewSolver(maxDepth, trace),
		trace:     trace,
		responder: responder,
		maxDepth:  maxDepth,
	}
}

// AddClaim stores a claim in the local state.
// This function shares a lock with PerformActions.
func (a *Agent) AddClaim(claim Claim) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.game.Put(claim)
}

// PerformActions iterates the game & performs all of the next actions.
// Note: PerformActions & AddClaim share a lock so the responder cannot
// call AddClaim on the same thread.
func (a *Agent) PerformActions() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, pair := range a.game.ClaimPairs() {
		a.move(pair.claim, pair.parent)
	}
}

// move determines & executes the next move given a claim pair
func (a *Agent) move(claim, parent Claim) {
	move, err := a.solver.NextMove(claim, parent)
	if err != nil || move == nil {
		return
	}
	// TODO(CLI-4123): Don't send duplicate responses
	// if a.game.IsDuplicate(move) {
	// 	return
	// }
	a.responder.Respond(context.TODO(), *move)
}
