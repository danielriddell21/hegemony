package sim

import "math"

func SpreadSpawns(width, height, n int) []Point {
	pts := make([]Point, 0, n)
	used := make(map[Point]struct{}, n)
	cx := float64(width-1) / 2
	cy := float64(height-1) / 2
	radius := 0.42 * math.Min(float64(width), float64(height))
	for i := range n {
		angle := 2 * math.Pi * float64(i) / float64(n)
		p := Point{
			X: clamp(int(cx+radius*math.Cos(angle)+0.5), 0, width-1),
			Y: clamp(int(cy+radius*math.Sin(angle)+0.5), 0, height-1),
		}
		p = nudgeFree(p, used, width, height)
		used[p] = struct{}{}
		pts = append(pts, p)
	}
	return pts
}

func nudgeFree(p Point, used map[Point]struct{}, width, height int) Point {
	if _, taken := used[p]; !taken {
		return p
	}
	for r := 1; r < width+height; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				q := Point{X: p.X + dx, Y: p.Y + dy}
				if q.X < 0 || q.X >= width || q.Y < 0 || q.Y >= height {
					continue
				}
				if _, taken := used[q]; !taken {
					return q
				}
			}
		}
	}
	return p
}

func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}
