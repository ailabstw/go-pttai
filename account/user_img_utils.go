// Copyright 2018 The go-pttai Authors
// This file is part of the go-pttai library.
//
// The go-pttai library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-pttai library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-pttai library. If not, see <http://www.gnu.org/licenses/>.

package account

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/log"
	"github.com/nfnt/resize"
)

func NormalizeImage(theBytes []byte, maxWidth int, maxHeight int) (ImgType, uint16, uint16, []byte, error) {
	reader := bytes.NewReader(theBytes)
	img, format, err := image.Decode(reader)
	if err != nil {
		return ImgTypeJPEG, 0, 0, nil, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// gif
	if format == "gif" {
		return ImgTypeGIF, uint16(width), uint16(height), theBytes, nil
	}

	// normalize width / height
	normalizedWidth, normalizedHeight := normalizeSize(width, height, maxWidth, maxHeight)

	// resize and to jpeg
	normalizedImage := resize.Resize(uint(normalizedWidth), uint(normalizedHeight), img, resize.Lanczos3)

	newImage, err := imgWithMask(normalizedImage)
	if err != nil {
		return ImgTypeJPEG, 0, 0, nil, err
	}
	newBytes, err := imgToPNG(newImage)
	if err != nil {
		return ImgTypeJPEG, 0, 0, nil, err
	}

	newBounds := newImage.Bounds()
	newWidth := newBounds.Dx()
	newHeight := newBounds.Dy()

	log.Debug("NormalizeImage", "normalizedWidth", normalizedWidth, "normalizedHeight", normalizedHeight, "newWidth", newWidth, "newHeight", newHeight, "newBytes", len(newBytes))

	return ImgTypePNG, uint16(newWidth), uint16(newHeight), newBytes, nil
}

func imgWithMask(img image.Image) (image.Image, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	offsetWidth := (MaxProfileImgWidth - width) / 2
	offsetHeight := (MaxProfileImgHeight - height) / 2

	dst := image.NewRGBA(image.Rect(0, 0, MaxProfileImgWidth, MaxProfileImgHeight))
	p := image.Point{(MaxProfileImgWidth + 1) / 2, (MaxProfileImgHeight + 1) / 2}
	r := MaxProfileImgHeight / 2

	rect := image.Rect(offsetWidth, offsetHeight, offsetWidth+width-1, offsetHeight+height-1)

	draw.DrawMask(dst, rect, img, image.ZP, &circle{p, r}, rect.Min, draw.Over)

	return dst, nil
}

func imgToJPEG(img image.Image) ([]byte, error) {
	buffer := &bytes.Buffer{}
	err := jpeg.Encode(buffer, img, nil)
	if err != nil {
		return nil, err
	}

	theBytes := common.CloneBytes(buffer.Bytes())

	return theBytes, nil
}

func imgToPNG(img image.Image) ([]byte, error) {
	buffer := &bytes.Buffer{}
	err := png.Encode(buffer, img)
	if err != nil {
		return nil, err
	}

	theBytes := common.CloneBytes(buffer.Bytes())

	return theBytes, nil
}

func normalizeSize(width int, height int, maxWidth int, maxHeight int) (int, int) {
	if width == height { // XXX hack for width == height
		return maxWidth, maxHeight
	}

	newWidth := width
	newHeight := height
	if newWidth > maxWidth {
		newHeight = height * maxWidth / newWidth
		newWidth = maxWidth
	}
	log.Debug("normalizeSize: after width", "width", width, "height", height, "newWidth", newWidth, "newHeight", newHeight)

	if newHeight > maxHeight {
		newWidth = width * maxHeight / height
		newHeight = maxHeight
	}

	log.Debug("normalizeSize: after height", "width", width, "height", height, "newWidth", newWidth, "newHeight", newHeight)

	return newWidth, newHeight
}
