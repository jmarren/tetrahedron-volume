package main

import (
	"bufio"
	"cmp"
	"fmt"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"time"
)

const (
	floatPattern = `[-+]?[0-9]*\.?[0-9]+`
	intPattern   = `\s[-+]?[0-9]+`
)

type DataPoint struct {
	location []float64
	nVal     int
	index    int
}
type smallestTetra struct {
	points          []*DataPoint
	volume          float64
	originalIndices []int
}
type Data struct {
	points        []*DataPoint
	smallestTetra smallestTetra
}

func reportTime(startTime time.Time) {
	totalTime := time.Since(startTime)
	fmt.Printf("\nTotal Execution Time: %v\n", totalTime)
}

func main() {
	begin := time.Now()
	defer reportTime(begin)

	// Start the pprof HTTP server in a separate goroutine for profiling performance
	go func() {
		log.Println("Starting pprof server on :6060")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	file, err := os.Open("../data/points_small.txt")
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)

	dataPoints, creationErr := CreateNewDataSet()
	if creationErr != nil {
		panic(creationErr)
	}

	dataPoints.ParsePoints(scanner)

	nValCmp := func(a, b *DataPoint) int {
		return cmp.Compare(a.nVal, b.nVal)
	}
	slices.SortFunc(dataPoints.points, nValCmp)

	for i := 0; i < 5; i++ {
		fmt.Println(dataPoints.points[i])
	}

	for i := len(dataPoints.points) - 1; i > len(dataPoints.points)-5; i-- {
		fmt.Println(dataPoints.points[i])
	}

	numCores := runtime.NumCPU() * 3
	fmt.Printf("Number of Cores on Machine: %v", numCores)

	dataPoints.findSmallest()

	fmt.Println("\nsmallest Tetrahedron Volume: ", dataPoints.smallestTetra.volume)
	fmt.Println("smallest Tetrahedron Points: ",
		[]DataPoint{
			*dataPoints.smallestTetra.points[0],
			*dataPoints.smallestTetra.points[1],
			*dataPoints.smallestTetra.points[2],
			*dataPoints.smallestTetra.points[3],
		})
	fmt.Println("smallest Tetrahedron Original Indices: ", dataPoints.smallestTetra.originalIndices)

	fmt.Println("\n------------- Testing Result -----------")
	testVolume := findVolume(dataPoints.smallestTetra.points[0].location, dataPoints.smallestTetra.points[1].location, dataPoints.smallestTetra.points[2].location, dataPoints.smallestTetra.points[3].location)
	fmt.Printf("\nVolume: %v", testVolume)

	fmt.Println("\n\n----------------- Testing Result with Hard-Coded Values for points_small.txt result -----------------")
	testVolume2 := findVolume([]float64{365.28, 374.98, 14.8}, []float64{432.13, 109.19, 264.16}, []float64{384.36, 176.25, 56.62}, []float64{300.7, 404.12, 257.92})
	fmt.Printf("\nVolume: %v", testVolume2)

	fmt.Println("\n\n----------------- Testing Result with Hard-Coded Values for points_large.txt result -----------------")
	testVolume3 := findVolume([]float64{276.81, 69.17, 142.37}, []float64{134.53, 292.87, 385.94}, []float64{88.74, 442.01, 395.32}, []float64{156.04, 326.98, 265.29})
	fmt.Printf("\nVolume: %v", testVolume3)
}

func CreateNewDataSet() (*Data, error) {
	d := Data{}
	return &d, nil
}

func (d *Data) findSmallest() {
	target := 100
	d.smallestTetra.volume = math.MaxFloat64

	for i := 0; i < len(d.points)-3; i++ {
		for j := i + 1; j < len(d.points)-2; j++ {
			if d.points[j].nVal < target-d.points[i].nVal {
				for k := j + 1; k < len(d.points)-1; k++ {
					if d.points[k].nVal < target-d.points[i].nVal-d.points[j].nVal {
						for m := k + 1; m < len(d.points); m++ {
							total := d.points[i].nVal + d.points[j].nVal + d.points[k].nVal + d.points[m].nVal
							if total == 100 {
								volume := findVolume(d.points[i].location, d.points[j].location, d.points[k].location, d.points[m].location)
								if volume < d.smallestTetra.volume {
									d.smallestTetra.volume = volume
									d.smallestTetra.points = []*DataPoint{d.points[i], d.points[j], d.points[k], d.points[m]}
									d.smallestTetra.originalIndices = []int{d.points[i].index, d.points[j].index, d.points[k].index, d.points[m].index}
								}
							} else if total > 100 {
								break
							}
						}
					} else {
						break
					}
				}
			} else {
				break
			}
		}
	}
}

func (d *Data) ParsePoints(scanner *bufio.Scanner) {
	re := regexp.MustCompile(floatPattern)
	lineNumber := -1

	for scanner.Scan() {
		point := []float64{}
		line := scanner.Text()
		count := 0
		lineNumber++

		for _, match := range re.FindAllString(line, -1) {
			var dataPoint DataPoint
			count++

			if count == 4 {
				int, err := strconv.Atoi(match)
				dataPoint.nVal = int
				dataPoint.location = point
				dataPoint.index = lineNumber
				d.points = append(d.points, &dataPoint)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				val, err := strconv.ParseFloat(match, 64)
				point = append(point, val)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func testFindVolume() {
	// # Example points
	A := []float64{1, 2, 3}
	B := []float64{2, 3, 4}
	C := []float64{1, 5, 1}
	D := []float64{4, 2, 3}

	// # Calculate the volume
	vol := findVolume(A, B, C, D)
	fmt.Println("The volume of the tetrahedron is ", vol)
}

func findVolume(p1 []float64, p2 []float64, p3 []float64, p4 []float64) float64 {
	AB := []float64{p2[0] - p1[0], p2[1] - p1[1], p2[2] - p1[1]}
	AC := []float64{p3[0] - p1[0], p3[1] - p1[1], p3[2] - p1[2]}
	AD := []float64{p4[0] - p1[0], p4[1] - p1[1], p4[2] - p1[2]}

	//  Direct calculation of the cross product components
	crossProductX := AB[1]*AC[2] - AB[2]*AC[1]
	crossProductY := AB[2]*AC[0] - AB[0]*AC[2]
	crossProductZ := AB[0]*AC[1] - AB[1]*AC[0]

	// # Dot product of AD with the cross product of AB and AC
	ScalarTripleProduct := (AD[0] * crossProductX) + (AD[1] * crossProductY) + (AD[2] * crossProductZ)

	// # The volume of the tetrahedron
	volume := math.Abs(ScalarTripleProduct) / 6.0
	return volume
}
