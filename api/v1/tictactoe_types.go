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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TicTacToeSpec defines the desired state of TicTacToe
type TicTacToeSpec struct {

	// User's move
	// +kubebuilder:validation:Pattern:=^[a-cA-C][1-3]$
	Move string `json:"move"`
}

// TicTacToeStatus defines the observed state of TicTacToe
type TicTacToeStatus struct {

	// Specifies whose turn is it to play
	Turn Identifier `json:"turn,omitempty"`

	// Game state table
	Table string `json:"table,omitempty"`

	// Indicates who won the game
	Winner string `json:"winner,omitempty"`

	// Moves that have been made
	MoveHistory []string `json:"moveHistory,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TicTacToe is the Schema for the tictactoes API
type TicTacToe struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TicTacToeSpec   `json:"spec,omitempty"`
	Status TicTacToeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TicTacToeList contains a list of TicTacToe
type TicTacToeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TicTacToe `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TicTacToe{}, &TicTacToeList{})
}
