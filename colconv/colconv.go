package colconv

// RgbToYUV converts a color from RGB colorspace to YUV colorspace
func RgbToYUV(r uint8, g uint8, b uint8) (y float64, u float64, v float64) {
	// define constants
	var rconst = 0.299
	var gconst = 0.587
	var bconst = 0.114
	var uMax = 0.436
	var vMax = 0.615

	rf := float64(r) / 255
	gf := float64(g) / 255
	bf := float64(b) / 255

	y = rconst*rf + gconst*gf + bconst*bf

	y = clampAbsf(y, 1)

	u = 0.492 * (float64(bf) - y)
	v = 0.877 * (float64(rf) - y)

	u = clampAbsf(u, uMax)
	v = clampAbsf(v, vMax)

	return y, u, v
}

// YuvToRGB converts a color from YUV colorspace to RGB colorspace
func YuvToRGB(y float64, u float64, v float64) (r uint8, g uint8, b uint8) {

	r = uint8((y + 1.14*v) * 255)
	g = uint8((y - 0.395*u - 0.581*v) * 255)
	b = uint8((y + 2.033*u) * 255)

	return r, g, b
}

func clampAbsf(f float64, limit float64) float64 {
	if limit < 0 {
		limit *= -1
	}

	if f > limit {
		f = limit
	} else if f < -limit {
		f = -limit
	}
	return f
}
