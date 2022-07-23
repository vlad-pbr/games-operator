/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gamesv1 "github.com/vlad-pbr/games-operator/api/v1"
)

// TicTacToeReconciler reconciles a TicTacToe object
type TicTacToeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type Column string

type Row string

type Slot struct {
	Column Column
	Row    Row
}

type SlotSymbol string

type Table map[Slot]SlotSymbol

type Move struct {
	Symbol SlotSymbol
	Slot   Slot
}

type Winner string

const (
	ColumnA Column = "A"
	ColumnB Column = "B"
	ColumnC Column = "C"
)

const (
	Row1 Row = "1"
	Row2 Row = "2"
	Row3 Row = "3"
)

const (
	SlotSymbolX     SlotSymbol = "X"
	SlotSymbolO     SlotSymbol = "O"
	SlotSymbolEmpty SlotSymbol = " "
)

var winCombinations = [][]Slot{
	{Slot{ColumnA, Row1}, Slot{ColumnA, Row2}, Slot{ColumnA, Row3}}, // column A filled
	{Slot{ColumnB, Row1}, Slot{ColumnB, Row2}, Slot{ColumnB, Row3}}, // column B filled
	{Slot{ColumnC, Row1}, Slot{ColumnC, Row2}, Slot{ColumnC, Row3}}, // column C filled
	{Slot{ColumnA, Row1}, Slot{ColumnB, Row1}, Slot{ColumnC, Row1}}, // row 1 filled
	{Slot{ColumnA, Row2}, Slot{ColumnB, Row2}, Slot{ColumnC, Row2}}, // row 2 filled
	{Slot{ColumnA, Row3}, Slot{ColumnB, Row3}, Slot{ColumnC, Row3}}, // row 3 filled
	{Slot{ColumnA, Row1}, Slot{ColumnB, Row2}, Slot{ColumnC, Row3}}, // backward diagonal filled
	{Slot{ColumnA, Row3}, Slot{ColumnB, Row2}, Slot{ColumnC, Row1}}, // forward diagonal filled
}

var possibleSlots = []Slot{
	{ColumnA, Row1},
	{ColumnA, Row2},
	{ColumnA, Row3},
	{ColumnB, Row1},
	{ColumnB, Row2},
	{ColumnB, Row3},
	{ColumnC, Row1},
	{ColumnC, Row2},
	{ColumnC, Row3},
}

//+kubebuilder:rbac:groups=games.vlad.io,resources=tictactoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=games.vlad.io,resources=tictactoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=games.vlad.io,resources=tictactoes/finalizers,verbs=update

func (s Slot) String() string {
	return fmt.Sprintf("%s%s", s.Column, s.Row)
}

func (m Move) String() string {
	return fmt.Sprintf("%s - %s", m.Slot, m.Symbol)
}

func (t *Table) get(s Slot) SlotSymbol {

	symbol, ok := (*t)[s]
	if !ok {
		symbol = SlotSymbolEmpty
	}

	return symbol
}

func (t *Table) set(s Slot, ss SlotSymbol) {

	if ss != SlotSymbolEmpty {
		(*t)[s] = ss
	} else if t.get(s) != SlotSymbolEmpty {
		delete(*t, s)
	}

}

func (t Table) String() string {

	out := fmt.Sprintf("   %s     %s     %s  \n", ColumnA, ColumnB, ColumnC)

	// iterate rows
	for _, rowPair := range []struct {
		Row
		bool
	}{{Row1, true}, {Row2, true}, {Row3, false}} {

		out += fmt.Sprintf("      |     |     \n%s", rowPair.Row)

		// iterate columns
		for _, columnPair := range []struct {
			Column
			bool
		}{{ColumnA, true}, {ColumnB, true}, {ColumnC, false}} {

			out += fmt.Sprintf("  %s  ", t.get(Slot{columnPair.Column, rowPair.Row}))

			if columnPair.bool {
				out += "|"
			} else {
				out += "\n"
			}

		}

		if rowPair.bool {
			out += " _____|_____|_____\n"
		} else {
			out += "      |     |     \n"
		}

	}

	return out
}

func (t *Table) read(ttt *gamesv1.TicTacToe) {

	if ttt.Status.Table != "" {

		split := strings.Split(ttt.Status.Table, "\n")

		for _, coordinate := range []struct {
			x int
			y int
			s Slot
		}{{2, 3, Slot{ColumnA, Row1}},
			{2, 9, Slot{ColumnB, Row1}},
			{2, 15, Slot{ColumnC, Row1}},
			{5, 3, Slot{ColumnA, Row2}},
			{5, 9, Slot{ColumnB, Row2}},
			{5, 15, Slot{ColumnC, Row2}},
			{8, 3, Slot{ColumnA, Row3}},
			{8, 9, Slot{ColumnB, Row3}},
			{8, 15, Slot{ColumnC, Row3}}} {

			t.set(coordinate.s, SlotSymbol(split[coordinate.x][coordinate.y]))
		}
	}
}

func (t *Table) getCurrentSymbol() SlotSymbol {

	counter := 0

	for _, ss := range *t {
		if ss == SlotSymbolX {
			counter++
		} else if ss == SlotSymbolO {
			counter--
		}
	}

	if counter == 0 {
		return SlotSymbolX
	} else {
		return SlotSymbolO
	}

}

func resolvePlayerSlot(ttt *gamesv1.TicTacToe) Slot {
	return Slot{Column(string(ttt.Spec.Move[0])), Row(string(ttt.Spec.Move[1]))}
}

func performPlayerMove(t Table, s Slot) (bool, Slot) {

	if t.get(s) != SlotSymbolEmpty {
		return false, s
	}

	return true, s
}

func performComputerMove(t Table, ss SlotSymbol) Slot {

	// init possibilities array
	availableSlots := []Slot{}

	// select all empty slots
	for _, slot := range possibleSlots {
		if t.get(slot) == SlotSymbolEmpty {
			availableSlots = append(availableSlots, slot)
		}
	}

	// pick random slot
	rand.Seed(time.Now().Unix())
	return availableSlots[rand.Intn(len(availableSlots))]
}

func resolveEndgame(t Table) (bool, bool) {

	// iterate possible win combinations
	for _, slots := range winCombinations {

		// store first symbol
		symbol := t.get(slots[0])
		win := true

		// do not match empty slots
		if symbol == SlotSymbolEmpty {
			continue
		}

		// iterate remaining slots
		for _, slot := range slots[1:] {

			// if mismatch - no win
			if t.get(slot) != symbol {
				win = false
				break
			}
		}

		// if all slots match - win
		if win {
			return true, false
		}
	}

	// if no win and all slots filled - stalemate
	if len(t) == 9 {
		return false, true
	}

	return false, false
}

func performMove(ttt *gamesv1.TicTacToe) {

	// read table
	table := make(Table)
	table.read(ttt)

	// get current symbol
	currentSymbol := table.getCurrentSymbol()

	// define move result vars
	var movePerformed bool
	var slot Slot

	// perform move
	if ttt.Status.Turn == gamesv1.IdentifierPlayer {
		if ttt.Spec.Move != "" {
			movePerformed, slot = performPlayerMove(table, resolvePlayerSlot(ttt))
		} else {
			movePerformed, slot = false, Slot{}
		}
	} else {
		movePerformed, slot = true, performComputerMove(table, currentSymbol)
	}

	if movePerformed {

		// store move in status
		table.set(slot, currentSymbol)
		ttt.Status.Table = table.String()
		ttt.Status.MoveHistory = append(ttt.Status.MoveHistory, Move{currentSymbol, slot}.String())

		// check for endgame
		win, stalemate := resolveEndgame(table)
		if win {
			ttt.Status.Winner = string(currentSymbol)
		} else if stalemate {
			ttt.Status.Winner = "Draw"
		} else {
			swapTurns(ttt)
		}
	}
}

func swapTurns(ttt *gamesv1.TicTacToe) {

	if ttt.Status.Turn == gamesv1.IdentifierPlayer {
		ttt.Status.Turn = gamesv1.IdentifierComputer
	} else {
		ttt.Status.Turn = gamesv1.IdentifierPlayer
	}

}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TicTacToeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// get current game from cluster
	ttt := &gamesv1.TicTacToe{}
	err := r.Client.Get(ctx, req.NamespacedName, ttt)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// do not reconcile if winner exists
	if ttt.Status.Winner != "" {
		return ctrl.Result{}, nil
	}

	// if table is empty - init new game
	if ttt.Status.Table == "" {

		// resolve first move
		if ttt.Spec.Move != "" {
			ttt.Status.Turn = gamesv1.IdentifierPlayer
		} else {
			ttt.Status.Turn = gamesv1.IdentifierComputer
		}
	}

	// if new move was performed - swap turns
	performMove(ttt)

	// update game status
	if err := r.Status().Update(ctx, ttt); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: false}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TicTacToeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gamesv1.TicTacToe{}).
		Complete(r)
}
