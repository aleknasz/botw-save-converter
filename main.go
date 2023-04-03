package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"

	"github.com/yargevad/filepathx"
)

const (
	WiiU     = 0x1E470000
	WiiU_B   = 0x1B470000
	Switch   = 0x0000471E
	Switch_B = 0x0000471B
)

var Headers = []uint16{0x24e2, 0x24EE, 0x2588, 0x29c0,
	0x3ef8, 0x471a, 0x471b, 0x471e}

var Versions = []string{"v1.0", "v1.1", "v1.2", "v1.3",
	"v1.3.3", "v1.4", "v1.5", "v1.6"}

var Items = []string{
	"Item", "Weap", "Armo", "Fire", "Norm", "IceA", "Elec", "Bomb", "Anci", "Anim",
	"Obj_", "Game", "Dm_N", "Dm_A", "Dm_E", "Dm_P", "FldO", "Gano", "Gian", "Grea",
	"KeyS", "Kokk", "Liza", "Mann", "Mori", "Npc_", "OctO", "Octa", "Octa", "arro",
	"Pict", "PutR", "Rema", "Site", "TBox", "TwnO", "Prie", "Dye0", "Dye1", "Map",
	"Play", "Oasi", "Cele", "Wolf", "Gata", "Ston", "Kaka", "Soji", "Hyru", "Powe",
	"Lana", "Hate", "Akka", "Yash", "Dung", "BeeH", "Boar", "Boko", "Brig", "DgnO"}

var Hashes = []uint32{
	0x7B74E117, 0x17E1747B, 0xD913B769, 0x69B713D9, 0xB666D246, 0x46D266B6, 0x021A6FF2,
	0xF26F1A02, 0xFF74960F, 0x0F9674FF, 0x8932285F, 0x5F283289, 0x3B0A289B, 0x9B280A3B,
	0x2F95768F, 0x8F76952F, 0x9C6CFD3F, 0x3FFD6C9C, 0xBBAC416B, 0x6B41ACBB, 0xCCAB71FD,
	0xFD71ABCC, 0xCBC6B5E4, 0xE4B5C6CB, 0x2CADB0E7, 0xE7B0AD2C, 0xA6EB3EF4, 0xF43EEBA6,
	0x21D4CFFA, 0xFACFD421, 0x22A510D1, 0xD110A522, 0x98D10D53, 0x530DD198, 0x55A22047,
	0x4720A255, 0xE5A63A33, 0x333AA6E5, 0xBEC65061, 0x6150C6BE, 0xBC118370, 0x708311BC,
	0x0E9D0E75, 0x750E9D0E}

type BotwSave struct {
	saveType     uint32
	sourceFolder string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func reverse(values []byte) []byte {
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
	return values
}

func readUInt32(source []byte, pos int) (uint32, []byte) {
	in := source[pos : pos+4]
	return binary.LittleEndian.Uint32(in), in
}

func readUInt16(source []byte, pos int) (uint16, []byte) {
	in := source[pos : pos+2]
	return binary.LittleEndian.Uint16(in), in
}

func writeReversed(target []byte, pos int, value []byte) {
	copy(target[pos:], reverse(value))
}

func ItemsContain(x []byte) bool {
	xx := string(x)
	for _, n := range Items {
		if xx == n {
			return true
		}
	}
	return false
}

func HashesContain(x uint32) bool {
	for _, n := range Hashes {
		if x == n {
			return true
		}
	}
	return false
}

func (state *BotwSave) Convert() {

	matches, err := filepathx.Glob(fmt.Sprintf("%s/**/*.sav", state.sourceFolder))
	check(err)

	for _, saveFilename := range matches {

		fmt.Printf("Processing %s\n", saveFilename)

		content, err := ioutil.ReadFile(saveFilename)
		check(err)

		contentLength := len(content) / 4
		var inPos = 0
		var pos = 0

		if strings.Contains(saveFilename, "trackblock") {
			inPos = 4
			_, rb2 := readUInt16(content, inPos)
			writeReversed(content, inPos, rb2)
			pos = 2
		}

		for ; pos < contentLength; pos += 1 {

			inPos = pos * 4
			i32, buffer := readUInt32(content, inPos)

			if HashesContain(i32) {
				writeReversed(content, inPos, buffer)
				pos += 1
			} else {
				if !ItemsContain(buffer) {
					writeReversed(content, inPos, buffer)
				} else {
					pos += 1
					for i := 0; i < 16; i += 1 {
						inPos = (pos + (i * 2)) * 4
						_, buffer := readUInt32(content, inPos)
						writeReversed(content, inPos, buffer)
					}

					pos += 30
				}
			}

		}

		ioutil.WriteFile(saveFilename, content, fs.ModePerm)

	}

	fmt.Printf("Finished converting %s save files to %s\n", state.SaveTypeName(false), state.SaveTypeName(true))
}

func (state *BotwSave) Load(sourceFolder string) {
	state.sourceFolder = sourceFolder

	header := fmt.Sprintf("%s/option.sav", sourceFolder)

	f, err := os.Open(header)
	check(err)

	defer f.Close()

	r4 := bufio.NewReader(f)
	b4, err := r4.Peek(4)
	check(err)

	state.saveType = binary.LittleEndian.Uint32(b4)
}

func (state *BotwSave) SaveTypeName(opposite bool) string {
	if state.saveType == WiiU || state.saveType == WiiU_B {
		if opposite {
			return "Switch"
		} else {
			return "Wii U"
		}
	} else if state.saveType == Switch || state.saveType == Switch_B {
		if opposite {
			return "Wii U"
		} else {
			return "Switch"
		}
	}
	return "unknown"
}

func main() {

	fmt.Println("Zelda: Breath of The Wild - save files converter")

	fmt.Print("Enter path to convert: ")

	var inputFolder string
	_, err := fmt.Scanln(&inputFolder)
	check(err)

	var saveFiles BotwSave

	saveFiles.Load(inputFolder)

	saveFiles.Convert()

}
