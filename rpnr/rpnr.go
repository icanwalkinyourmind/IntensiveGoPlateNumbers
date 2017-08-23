package rpnr

import (
	"strconv"
	"unsafe"

	"github.com/lazywei/go-opencv/opencv"
	"github.com/otiai10/gosseract"
)

func GetPlateNumber(filename string) (string, string) {
	image := opencv.LoadImage(filename, 0)
	if image == nil {
		return "", "Couldn`t open file \n"
	}
	defer image.Release()

	cascade := opencv.LoadHaarClassifierCascade("test.xml")
	if cascade == nil {
		return "", "Couldn`t find cacade file\n"
	}
	numbers := cascade.DetectObjects(image)
	if numbers == nil {
		return "", "Couldn`t detect plates\n"
	}
	//Getting the plate from the whole picture
	var plate *opencv.IplImage
	defer plate.Release()
	for _, value := range numbers {
		plate = opencv.Crop(image, value.X(), value.Y(), value.Width(), value.Height())
		opencv.Threshold(plate, plate, 120, 255, opencv.CV_THRESH_BINARY)
		//opencv.SaveImage("plate.png", plate, nil) //
	}
	copyPlate := plate.Clone()
	defer copyPlate.Release()

	//Getting all contours on the plate as Seq
	var point opencv.Point
	point.X = 0
	point.Y = 0
	borderSequance := copyPlate.FindContours(opencv.CV_RETR_LIST, opencv.CV_CHAIN_APPROX_SIMPLE, point)
	defer borderSequance.Release()
	numberOfBorders := 0
	copyBorderSequance := borderSequance
	defer copyBorderSequance.Release()

	for i := 0; borderSequance != nil; i++ {
		numberOfBorders++
		borderSequance = borderSequance.HNext()
	}

	//Getting all contours on the plate as slice (array)
	borderArray := make([]opencv.Rect, numberOfBorders)
	for i := 0; copyBorderSequance != nil; i++ {
		borderArray[i] = opencv.BoundingRect(unsafe.Pointer(copyBorderSequance))
		copyBorderSequance = copyBorderSequance.HNext()
	}

	//Sorting borders by the size
	for i := 0; i < numberOfBorders-1; i++ {
		for j := 0; j < numberOfBorders-i-1; j++ {
			if borderArray[j].Height()*borderArray[j].Width() > borderArray[j+1].Height()*borderArray[j+1].Width() {
				borderArray[j], borderArray[j+1] = borderArray[j+1], borderArray[j]
			}
		}
	}

	//Sorting borders by the pozition
	sortedBorders := borderArray[numberOfBorders-11 : numberOfBorders-2]
	for i := 0; i < 9; i++ {
		for j := 0; j < 9-i-1; j++ {
			if sortedBorders[j].X() > sortedBorders[j+1].X() {
				sortedBorders[j], sortedBorders[j+1] = sortedBorders[j+1], sortedBorders[j]
			}
		}
	}
	//Saving each character's border
	for i := 0; i < 9; i++ {
		symbol := opencv.Crop(plate, sortedBorders[i].X(), sortedBorders[i].Y(), sortedBorders[i].Width(), sortedBorders[i].Height())
		opencv.SaveImage("src/symbol"+strconv.Itoa(i)+".png", symbol, nil)
	}
	//Recognition of each character
	alpha := "ABEKMHOPCTYX0123456789"
	result := ""
	flag := false
	//os.Stderr.Close()
	for i := 0; i < 9; i++ {
		character := gosseract.Must(gosseract.Params{
			Src:       "src/symbol" + strconv.Itoa(i) + ".png",
			Languages: "eng",
		})
		flag = true
		if len(character) == 0 {
			continue
		}
		for i := 0; i < len(alpha); i++ {
			if character[0] == alpha[i] {
				character = character[0:1]
				result += character
				flag = false
				break
			}
		}
		if flag {
			result += "*"
		}

	}

	return result, ""
}
