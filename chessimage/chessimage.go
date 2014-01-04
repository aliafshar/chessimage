package chessimage

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	SQUARE_SIZE = 45
)

var pieces map[string]image.Image = make(map[string]image.Image)

var piece_names [12]string = [12]string{"r", "n", "b", "q", "k", "p", "R", "N", "B", "Q", "K", "P"}

func LoadPiece(name string) (image.Image, error) {
	reader, err := os.Open("images/" + name + ".png")
	if err != nil {
		return nil, err
	}
	m, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func InitPieces() error {
	for _, p := range piece_names {
		m, err := LoadPiece(p)
		if err != nil {
			return err
		}
		pieces[p] = m
	}
	return nil
}

type BoardConfig struct {
	SquareSize int
	BoardSize  int
}

func NewBoardConfig(SquareSize int) *BoardConfig {
	bc := &BoardConfig{SquareSize: SquareSize}
	bc.BoardSize = bc.SquareSize * 8
	return bc
}

type Position struct {
	pieces    [][]string
	active    string
	castling  string
	enpassant string
	halfmove  int
	fullmove  int
}

func GetPosition(fen string) (*Position, error) {
	parts := strings.Split(fen, " ")
	if len(parts) < 1 {
		return nil, errors.New("Bad fen")
	}
	fenpieces := strings.Split(parts[0], "/")
	if len(fenpieces) != 8 {
		return nil, errors.New("Bad fen")
	}

	pieces := make([][]string, 8)

	for i, rps := range fenpieces {
		rank := make([]string, 0)
		for _, c := range rps {
			p := string(c)
			count, err := strconv.Atoi(p)
			if err == nil {
				for k := 0; k < count; k++ {
					rank = append(rank, "_")
				}
			} else {
				rank = append(rank, p)
			}
		}
		if len(rank) != 8 {
			return nil, errors.New("Bad fen")
		}
		pieces[7-i] = rank
	}

	pos := &Position{}
	pos.pieces = pieces
	return pos, nil
}

func PrintBoard(pos *Position) {
	for _, rank := range pos.pieces {
		log.Println(rank)
	}
}

func DisplayBoard(bc *BoardConfig, pos *Position, w io.Writer) {
	m := image.NewRGBA(image.Rect(0, 0, bc.BoardSize, bc.BoardSize))
	black := color.RGBA{77, 109, 146, 255}
	white := color.RGBA{236, 236, 215, 255}
	draw.Draw(m, m.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)

	for r, fs := range pos.pieces {
		for f, ss := range fs {
			p := image.Point{f * bc.SquareSize, bc.BoardSize - r*bc.SquareSize - bc.SquareSize}
			s := image.Rect(p.X, p.Y, p.X+bc.SquareSize, p.Y+bc.SquareSize)
			log.Println(s)
			if (r+f)%2 == 1 {
				draw.Draw(m, s, &image.Uniform{white}, image.ZP, draw.Src)
			}
			if ss != "_" {
				draw.Draw(m, s, pieces[ss], image.ZP, draw.Over)
			}
		}
	}

	jpeg.Encode(w, m, &jpeg.Options{Quality: 100})

}

func main() {
	err := InitPieces()
	if err != nil {
		log.Fatal("Could not load pieces.")
	}
	bc := NewBoardConfig(SQUARE_SIZE)
	fen := "rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2"
	pos, err := GetPosition(fen)
	DisplayBoard(bc, pos, os.Stdout)
}

func init() {
	err := InitPieces()
	if err != nil {
		log.Println("Could not load pieces.", err)
	}
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	bc := NewBoardConfig(SQUARE_SIZE)
	fen := strings.Trim(r.URL.Path, "/")
	pos, err := GetPosition(fen)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, err)
	} else {
		w.Header().Set("Content-type", "image/jpeg")
		DisplayBoard(bc, pos, w)
	}
}
