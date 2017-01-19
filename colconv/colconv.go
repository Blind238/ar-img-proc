package colconv

// RgbToYUV converts a color from RGB colorspace to YUV colorspace
func RgbToYUV(r uint8, g uint8, b uint8) (y float32, u float32, v float32) {
	// define constants
	var rconst float32 = 0.299
	var gconst float32 = 0.587
	var bconst float32 = 0.114
	var uMax float32 = 0.436
	var vMax float32 = 0.615

	rf := float32(r) / 255
	gf := float32(g) / 255
	bf := float32(b) / 255

	y = rconst*rf + gconst*gf + bconst*bf

	y = clampAbsf(y, 1)

	u = 0.492 * (float32(bf) - y)
	v = 0.877 * (float32(rf) - y)

	u = clampAbsf(u, uMax)
	v = clampAbsf(v, vMax)

	return y, u, v
}

// YuvToRGB converts a color from YUV colorspace to RGB colorspace
func YuvToRGB(y float32, u float32, v float32) (r uint8, g uint8, b uint8) {

	r = uint8((y + 1.14*v) * 255)
	g = uint8((y - 0.395*u - 0.581*v) * 255)
	b = uint8((y + 2.033*u) * 255)

	return r, g, b
}

func clampAbsf(f float32, limit float32) float32 {
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
