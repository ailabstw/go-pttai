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

package service

import (
	"bytes"
	"image"
	"image/jpeg"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/log"
	"github.com/nfnt/resize"
)

func NormalizeImage(theBytes []byte) (MediaType, interface{}, []byte, error) {

	maxWidth := MaxUploadImageWidth
	maxHeight := MaxUploadImageHeight

	reader := bytes.NewReader(theBytes)
	img, format, err := image.Decode(reader)
	if err != nil {
		return MediaTypeJPEG, nil, nil, err
	}

	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// gif
	if format == "gif" {
		return MediaTypeGIF, &MediaDataGIF{Width: uint16(width), Height: uint16(height)}, theBytes, nil
	}

	// good width and height
	if width <= maxWidth && height <= maxHeight {
		if format == "png" {
			newBytes, err := imgToJPEG(img)
			if err != nil {
				return MediaTypeJPEG, nil, nil, err
			}
			theBytes = newBytes
		}
		return MediaTypeJPEG, &MediaDataJPEG{Width: uint16(width), Height: uint16(height)}, theBytes, nil
	}

	// normalize width / height
	normalizedWidth, normalizedHeight := normalizeSize(width, height, maxWidth, maxHeight)

	// resize and to jpeg
	newImage := resize.Resize(uint(normalizedWidth), uint(normalizedHeight), img, resize.Lanczos3)
	newBytes, err := imgToJPEG(newImage)
	if err != nil {
		return MediaTypeGIF, &MediaDataGIF{Width: uint16(width), Height: uint16(height)}, theBytes, nil
	}

	return MediaTypeJPEG, &MediaDataJPEG{Width: uint16(normalizedWidth), Height: uint16(normalizedHeight)}, newBytes, nil
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

func normalizeSize(width int, height int, maxWidth int, maxHeight int) (int, int) {
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
