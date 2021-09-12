package main

import (
	"flag"
	"fmt"
	"github.com/fogleman/gg"
	"log"
	"math"
	"os"
	"path"
	"sort"
)

// Constants
const (
	Debug             = false
	NumStrings        = 6
	NumFrets          = 25
	NumCanonicalNotes = 12
	MinOctave         = 2
	MaxOctave         = 6
	ScaleCutoff       = 10
	PngSquareLength   = 200
	FontRegular       = "/usr/share/fonts/TTF/IBMPlexMono-Regular.ttf"
	FontSizeRegular   = 96
	FontLight         = "/usr/share/fonts/TTF/IBMPlexMono-Light.ttf"
	FontSizeLight     = 70
	FontSizeTitle     = 50
)

func getOpenStringNotes() [6]string {
	return [6]string{"E", "B", "G", "D", "A", "E"}
}

func getOpenStringOctaves() [6]int {
	return [6]int{4, 3, 3, 3, 2, 2}
}

func getCanonicalNotes() [12]string {
	return [12]string{
		"C", "C#", "D", "D#",
		"E", "F", "F#", "G",
		"G#", "A", "A#", "B"}
}

func getMinorScaleSteps() [8]int {
	return [8]int{2, 1, 2, 2, 1, 2, 2, 0}
}

func getMajorScaleSteps() [8]int {
	return [8]int{2, 2, 1, 2, 2, 2, 1, 0}
}

type MusicalNote struct {
	note   string
	octave int
}

type FretboardCoordinate struct {
	stringId int
	fretId   int
}

type Fretboard struct {
	musicalNotes [NumStrings][NumFrets]MusicalNote
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}

func getFretboard() *Fretboard {
	var fretboard *Fretboard
	fretboard = new(Fretboard)

	openStringNotes := getOpenStringNotes()
	openStringOctaves := getOpenStringOctaves()
	canonicalNotes := getCanonicalNotes()

	for stringId := 0; stringId < NumStrings; stringId++ {
		openStringNote := openStringNotes[stringId]
		openStringOctave := openStringOctaves[stringId]
		openStringCanonicalId := indexOf(openStringNote, canonicalNotes[:])

		for fretId := 0; fretId < NumFrets; fretId++ {
			fretboard.musicalNotes[stringId][fretId] = MusicalNote{
				canonicalNotes[(openStringCanonicalId+fretId)%NumCanonicalNotes],
				openStringOctave + (openStringCanonicalId+fretId)/NumCanonicalNotes,
			}
		}
	}
	return fretboard
}

func printFretboard(fretboard *Fretboard) {
	for stringId := 0; stringId < NumStrings; stringId++ {
		for fretId := 0; fretId < NumFrets; fretId++ {
			musicalNote := fretboard.musicalNotes[stringId][fretId]
			fmt.Print(musicalNote.note, musicalNote.octave, "\t")
		}
		fmt.Println()
	}
}

func getMusicalNoteCoordinates(musicalNote MusicalNote, fretboard *Fretboard) []FretboardCoordinate {
	var result []FretboardCoordinate

	for stringId := 0; stringId < NumStrings; stringId++ {
		for fretId := 0; fretId < NumFrets; fretId++ {
			if fretboard.musicalNotes[stringId][fretId] == musicalNote {
				result = append(result, FretboardCoordinate{stringId, fretId})
			}
		}
	}

	return result
}

func getMusicalNoteToCoordinatesMap(fretboard *Fretboard) map[MusicalNote][]FretboardCoordinate {
	result := make(map[MusicalNote][]FretboardCoordinate)
	canonicalNotes := getCanonicalNotes()

	for octave := MinOctave; octave <= MaxOctave; octave++ {
		for _, canonicalNote := range canonicalNotes {
			musicalNote := MusicalNote{canonicalNote, octave}
			result[musicalNote] = getMusicalNoteCoordinates(musicalNote, fretboard)
		}
	}

	return result
}

func getScale(rootMusicalNote MusicalNote, steps []int) []MusicalNote {
	canonicalNotes := getCanonicalNotes()
	rootNoteCanonicalId := indexOf(rootMusicalNote.note, canonicalNotes[:])
	rootNoteOctave := rootMusicalNote.octave

	var result []MusicalNote
	offset := 0
	for _, step := range steps {
		octave := rootNoteOctave + (rootNoteCanonicalId+offset)/NumCanonicalNotes
		musicalNote := MusicalNote{canonicalNotes[(rootNoteCanonicalId+offset)%NumCanonicalNotes], octave}
		result = append(result, musicalNote)

		offset += step
	}
	return result
}

func getAllFretboardSequences(scale []MusicalNote, fretboard *Fretboard) [][]FretboardCoordinate {
	var result [][]FretboardCoordinate = [][]FretboardCoordinate{}
	numNotes := len(scale)

	if numNotes > 0 {
		musicalNote, subScale := scale[0], scale[1:]
		musicalNoteCoords := getMusicalNoteCoordinates(musicalNote, fretboard)

		if musicalNoteCoords != nil {
			if numNotes == 1 {
				for _, musicalNoteCoord := range musicalNoteCoords {
					var wrap []FretboardCoordinate
					wrap = []FretboardCoordinate{musicalNoteCoord}
					result = append(result, wrap)
				}
			} else {
				var subResults [][]FretboardCoordinate = getAllFretboardSequences(subScale, fretboard)
				for _, musicalNoteCoord := range musicalNoteCoords {
					for _, subResult := range subResults {
						var compositeResult []FretboardCoordinate = append([]FretboardCoordinate{musicalNoteCoord}, subResult...)
						result = append(result, compositeResult)
					}
				}
			}
		}
	}

	return result
}

func manhattanDistance(x1 float64, y1 float64, x2 float64, y2 float64) float64 {
	return math.Abs(x2-x1) + math.Abs(y2-y1)
}

func euclideanDistance(x1 float64, y1 float64, x2 float64, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

func clusterDistance(fretboardSequence []FretboardCoordinate) float64 {
	meanStringId := 0.0
	meanFretId := 0.0
	for i, coord := range fretboardSequence {
		i := float64(i)
		meanStringId = (i*meanStringId + float64(coord.stringId)) / (i + 1)
		meanFretId = (i*meanFretId + float64(coord.fretId)) / (i + 1)
	}

	score := 0.0
	for _, coord := range fretboardSequence {
		score += euclideanDistance(float64(coord.stringId), float64(coord.fretId), meanStringId, meanFretId)
	}
	return score
}

type ScoredFretboardSequence struct {
	sequence []FretboardCoordinate
	score    float64
}

func isConsistent(fretboardSequence []FretboardCoordinate) bool {
	numCoords := len(fretboardSequence)

	for i := 1; i < numCoords; i++ {
		a, b := fretboardSequence[i-1], fretboardSequence[i]
		if a.stringId < b.stringId {
			return false
		}
	}

	return true
}

func scoreAndSortFretboardSequences(fretboardSequences [][]FretboardCoordinate) []ScoredFretboardSequence {
	var scoredFretboardSequence []ScoredFretboardSequence

	for _, fretboardSequence := range fretboardSequences {

		if isConsistent(fretboardSequence) {
			score := clusterDistance(fretboardSequence)
			scoredFretboardSequence = append(scoredFretboardSequence, ScoredFretboardSequence{fretboardSequence, score})
		}
	}

	sort.Slice(scoredFretboardSequence, func(i, j int) bool {
		return scoredFretboardSequence[i].score < scoredFretboardSequence[j].score
	})

	if len(scoredFretboardSequence) <= ScaleCutoff {
		return scoredFretboardSequence
	} else {
		return scoredFretboardSequence[:ScaleCutoff]
	}

}

func saveFretboardSequence(fretboardSequence ScoredFretboardSequence, fretboard *Fretboard, fname string) {
	width := (NumFrets + 2) * PngSquareLength
	height := (NumStrings + 2) * PngSquareLength

	dc := gg.NewContext(width, height)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	for stringId := 1; stringId <= NumStrings+1; stringId++ {
		stringId := float64(stringId)
		dc.DrawLine(0.0, stringId*PngSquareLength, float64(width), stringId*PngSquareLength)
		dc.Stroke()
	}
	for fretId := 1; fretId <= NumFrets+1; fretId++ {
		fretId := float64(fretId)
		dc.DrawLine(fretId*PngSquareLength, 0, fretId*PngSquareLength, float64(height))
		dc.Stroke()
	}

	if err := dc.LoadFontFace(FontRegular, FontSizeRegular); err != nil {
		panic(err)
	}

	for stringId := 1; stringId <= NumStrings; stringId++ {
		stringId := float64(stringId)
		dc.SetRGB(0.5, 0.5, 0.5)
		dc.DrawRectangle(0, stringId*PngSquareLength, PngSquareLength, PngSquareLength)
		dc.DrawRectangle((NumFrets+1)*PngSquareLength, stringId*PngSquareLength, PngSquareLength, PngSquareLength)
		dc.Fill()
		dc.SetRGB(0.9, 0.9, 0.6)
		dc.DrawRectangle(PngSquareLength, stringId*PngSquareLength, PngSquareLength, PngSquareLength)
		dc.DrawRectangle((NumFrets/2 + 1) * PngSquareLength, stringId*PngSquareLength, PngSquareLength, PngSquareLength)
		dc.DrawRectangle(NumFrets * PngSquareLength, stringId*PngSquareLength, PngSquareLength, PngSquareLength)
		dc.Fill()
		dc.SetRGB(0, 0, 0)
		dc.DrawStringAnchored(fmt.Sprintf("%d", int(stringId)), float64(PngSquareLength)/2, (stringId+0.5)*PngSquareLength, 0.5, 0.5)
		dc.DrawStringAnchored(fmt.Sprintf("%d", int(stringId)), (NumFrets+1.5)*PngSquareLength, (stringId+0.5)*PngSquareLength, 0.5, 0.5)
	}

	for fretId := 1; fretId <= NumFrets; fretId++ {
		fretId := float64(fretId)
		dc.SetRGB(0.5, 0.5, 0.5)
		dc.DrawRectangle(fretId*PngSquareLength, 0, PngSquareLength, PngSquareLength)
		dc.DrawRectangle(fretId*PngSquareLength, (NumStrings+1)*PngSquareLength, PngSquareLength, PngSquareLength)
		dc.Fill()
		dc.SetRGB(0, 0, 0)
		if fretId > 1 {
			dc.DrawStringAnchored(fmt.Sprintf("%d", int(fretId)-1), (fretId+0.5)*PngSquareLength, float64(PngSquareLength)/2, 0.5, 0.5)
			dc.DrawStringAnchored(fmt.Sprintf("%d", int(fretId)-1), (fretId+0.5)*PngSquareLength, (NumStrings+1.5)*PngSquareLength, 0.5, 0.5)
		}
	}

	for i, fretboardCoord := range fretboardSequence.sequence {
		dc.DrawStringAnchored(fmt.Sprintf("%d", i+1), (float64(fretboardCoord.fretId)+1.5)*PngSquareLength, (float64(fretboardCoord.stringId)+1.25)*PngSquareLength, 0.5, 0.5)
	}

	if err := dc.LoadFontFace(FontLight, FontSizeLight); err != nil {
		panic(err)
	}

	dc.SetRGB(0.75, 0.75, 0.75)
	for stringId := 0; stringId < NumStrings; stringId++ {
		for fretId := 0; fretId < NumFrets; fretId++ {
			musicalNote := fretboard.musicalNotes[stringId][fretId]
			dc.DrawStringAnchored(fmt.Sprintf("[%s%d]", musicalNote.note, musicalNote.octave), (float64(fretId)+1.5)*PngSquareLength, (float64(stringId)+1.75)*PngSquareLength, 0.5, 0.5)
		}
	}

	dc.SetRGB(0, 0, 0)
	for _, fretboardCoord := range fretboardSequence.sequence {
		musicalNote := fretboard.musicalNotes[fretboardCoord.stringId][fretboardCoord.fretId]
		dc.DrawStringAnchored(fmt.Sprintf("[%s%d]", musicalNote.note, musicalNote.octave), (float64(fretboardCoord.fretId)+1.5)*PngSquareLength, (float64(fretboardCoord.stringId)+1.75)*PngSquareLength, 0.5, 0.5)
	}

	if err := dc.LoadFontFace(FontLight, FontSizeTitle); err != nil {
		panic(err)
	}

	dirName := path.Dir(fname)
	dc.DrawStringAnchored(path.Base(dirName), float64(PngSquareLength)/2, float64(PngSquareLength)/2, 0.5, 0.5)
	os.Mkdir(dirName, 0755)
	dc.SavePNG(fname)
}

func generateAllSequences(fretboard *Fretboard, scaleType string, scaleOctave int) {

	var steps [8]int
	if scaleType == "min" {
		steps = getMinorScaleSteps()
	} else if scaleType == "maj" {
		steps = getMajorScaleSteps()
	}

	for _, rootNote := range getCanonicalNotes() {

		musicalRootNote := MusicalNote{rootNote, scaleOctave}
		scale := getScale(musicalRootNote, steps[:])
		scaleFretboardSequences := getAllFretboardSequences(scale, fretboard)
		if len(scaleFretboardSequences) > 0 {
			scaleScoredFretboardSequences := scoreAndSortFretboardSequences(scaleFretboardSequences)

			if Debug {
				for i, musicalNote := range scale {
					fmt.Printf("%d: %s%d\n", i, musicalNote.note, musicalNote.octave)
				}

				for _, scaleScoredFretboardSequence := range scaleScoredFretboardSequences {
					fmt.Printf("S = %.1f: ", scaleScoredFretboardSequence.score)
					for _, coord := range scaleScoredFretboardSequence.sequence {
						fmt.Printf("(%d %d)", coord.stringId, coord.fretId)
					}
					fmt.Println()
				}
			}

			for i, fretboardSequence := range scaleScoredFretboardSequences {
				saveFretboardSequence(fretboardSequence, fretboard, fmt.Sprintf("png/%s%d%s/%d.png", rootNote, scaleOctave, scaleType, i))
			}
		}
	}
}

func main() {
	var scaleOctave int
	flag.IntVar(&scaleOctave, "octave", 0, "2 <= octave <= 6")
	flag.Parse()
	if scaleOctave < 2 || scaleOctave > 6 {
		log.Print("Octave needs to be in interval [2,...,6]")
		return
	}

	os.Mkdir("png", 0755)

	fretboard := getFretboard()
	if Debug {
		printFretboard(fretboard)
	}

	generateAllSequences(fretboard, "min", scaleOctave)
	generateAllSequences(fretboard, "maj", scaleOctave)
}
