Games operator
==============

Wouldn't it be just ridiculous to take a production-grade infrastructure tool and fully implement an entire controller for it just to play some basic games?... Would it?

# Games

This version of the operator offers the following games:

## Tic-Tac-Toe

To start a new Player vs Computer game, create the following object in your namespace:
```yaml
apiVersion: games.vlad.io/v1
kind: TicTacToe
metadata:
  name: tictactoe-sample
spec: {}
```

You are now able to play the game by running `kubectl edit ttt tictactoe-sample`, observing the table and performing a move by editing the `.spec.move` field like so:
```yaml
apiVersion: games.vlad.io/v1
kind: TicTacToe
metadata:
  name: tictactoe-sample
spec:
  move: B2
  pvp: false
status:
  moveHistory:
  - C3 - X
  table: |2+

       A     B     C
          |     |
    1     |     |
     _____|_____|_____
          |     |
    2     |     |
     _____|_____|_____
          |     |
    3     |     |  X
          |     |

  turn: Player
```

If you wish to play against another player in your namespace, create the initial object with the `.spec.pvp` field set to `true`.