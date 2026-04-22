package game

import "strings"

func BoardSymbols(fen string, perspective Side) ([8][8]string, error) {
	var board [8][8]string
	if _, err := gameFromFEN(fen); err != nil {
		return board, err
	}
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			sqName, _ := ParseSquareName(file, rank)
			piece, _, ok, err := PieceAt(fen, sqName)
			if err != nil {
				return board, err
			}
			if !ok {
				board[rank][file] = " "
				continue
			}
			board[rank][file] = strings.ToUpper(string(piece))
			if piece >= 'a' && piece <= 'z' {
				board[rank][file] = strings.ToLower(board[rank][file])
			}
		}
	}
	if perspective == White {
		return board, nil
	}
	return flipBoard(board), nil
}

func flipBoard(src [8][8]string) [8][8]string {
	var dst [8][8]string
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			dst[rank][file] = src[7-rank][7-file]
		}
	}
	return dst
}
