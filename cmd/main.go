package main

import (
	"bufio"
	"cmp"
	"fmt"
	"math"
	_ "net/http/pprof"
	"os"
	"regexp"
	"slices"
	"strconv"
	"time"
)

const floatPattern = `[-+]?[0-9]*\.?[0-9]+`

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
	// Save the start time of the program and defer the function to report the total execution time at completion
	startTime := time.Now()
	defer reportTime(startTime)

	// Open the file containing the points_small data
	pointsSmallFile, err := os.Open("../data/points_small.txt")
	if err != nil {
		panic(err)
	}
	pointsLargeFile, err := os.Open("../data/points_large.txt")
	if err != nil {
		panic(err)
	}

	// create a scanner to read the file
	pointsSmallScanner := bufio.NewScanner(pointsSmallFile)
	pointsLargeScanner := bufio.NewScanner(pointsLargeFile)

	// Create a new data set to hold the points in a Data struct
	pointsSmall, creationErr := CreateNewDataSet()
	if creationErr != nil {
		panic(creationErr)
	}
	pointsLarge, creationErr := CreateNewDataSet()
	if creationErr != nil {
		panic(creationErr)
	}

	// Parse the points from the file and add them to the dataPoints.points slice
	// include the index of the point in the original file in dataPoints.points.index for later reference
	pointsSmall.ParsePoints(pointsSmallScanner)
	pointsLarge.ParsePoints(pointsLargeScanner)

	// Sort the dataPoints.points slice by the nVal field in ascending order
	// so that we can later skip over points that will not be able to create a valid tetrahedron
	nValCmp := func(a, b *DataPoint) int {
		return cmp.Compare(a.nVal, b.nVal)
	}
	slices.SortFunc(pointsSmall.points, nValCmp)
	slices.SortFunc(pointsLarge.points, nValCmp)

	// Print the first and last 5 points in the dataPoints.points slice (to verify they are sorted correctly)
	// for i := 0; i < 5; i++ {
	// 	fmt.Println(pointsSmall.points[i])
	// 	fmt.Println(pointsLarge.points[i])
	// }
	// for i := len(pointsSmall.points) - 1; i > len(pointsSmall.points)-5; i-- {
	// 	fmt.Println(pointsSmall.points[i])
	// 	fmt.Println(pointsLarge.points[i])
	// }

	// use the findSmallest method to find valid tetrahedrons with a total of 100
	// then compare them to the current smallest tetrahedron found
	// save the smallest tetrahedron found in the pointsSmall.smallestTetra struct
	pointsSmall.findSmallest()
	pointsLarge.findSmallest()

	// Print the results for points_small.txt
	fmt.Println("\n=============== Results ==================")
	fmt.Println("\n\n\n--------------- points_small.txt -----------")
	fmt.Println("\nsmallest Tetrahedron Volume: ", pointsSmall.smallestTetra.volume)
	fmt.Println("smallest Tetrahedron Points: ",
		[]DataPoint{
			*pointsSmall.smallestTetra.points[0],
			*pointsSmall.smallestTetra.points[1],
			*pointsSmall.smallestTetra.points[2],
			*pointsSmall.smallestTetra.points[3],
		})
	finalAnswerPointsSmall := pointsSmall.smallestTetra.originalIndices
	slices.Sort(finalAnswerPointsSmall)

	fmt.Println("\n\n ************************ points_small.txt smallest valid tetrahedron indicies: ", finalAnswerPointsSmall, "***************************")

	fmt.Println("\n............. Testing Result .............")
	testVolume := findVolume(pointsSmall.smallestTetra.points[0].location, pointsSmall.smallestTetra.points[1].location, pointsSmall.smallestTetra.points[2].location, pointsSmall.smallestTetra.points[3].location)
	fmt.Printf("\nVolume: %v", testVolume)

	fmt.Println("\n\n.................. Testing Result with Hard-Coded Values for points_small.txt result ..................")
	testVolume2 := findVolume([]float64{365.28, 374.98, 14.8}, []float64{432.13, 109.19, 264.16}, []float64{384.36, 176.25, 56.62}, []float64{300.7, 404.12, 257.92})
	fmt.Printf("\nVolume: %v", testVolume2)

	fmt.Println("\n\n\n -------------------- points_large.txt -----------------")

	fmt.Println("\nsmallest Tetrahedron Volume: ", pointsLarge.smallestTetra.volume)
	fmt.Println("smallest Tetrahedron Points: ",
		[]DataPoint{
			*pointsLarge.smallestTetra.points[0],
			*pointsLarge.smallestTetra.points[1],
			*pointsLarge.smallestTetra.points[2],
			*pointsLarge.smallestTetra.points[3],
		})
	finalAnswerPointsLarge := pointsLarge.smallestTetra.originalIndices
	slices.Sort(finalAnswerPointsLarge)

	fmt.Println("\n\n ************************ points_large.txt smallest valid tetrahedron indicies: ", finalAnswerPointsLarge, "***************************")

	fmt.Println("\n...... Testing Result .......")
	testVolumePointsLarge := findVolume(pointsLarge.smallestTetra.points[0].location, pointsLarge.smallestTetra.points[1].location, pointsLarge.smallestTetra.points[2].location, pointsLarge.smallestTetra.points[3].location)
	fmt.Printf("\nVolume: %v", testVolumePointsLarge)

	fmt.Println("\n\n............ Testing Result with Hard-Coded Values for points_large.txt result ..................")
	testVolumePointsLargeHardCoded := findVolume([]float64{276.81, 69.17, 142.37}, []float64{134.53, 292.87, 385.94}, []float64{88.74, 442.01, 395.32}, []float64{156.04, 326.98, 265.29})
	fmt.Printf("\nVolume: %v", testVolumePointsLargeHardCoded)

	fmt.Println("\n\n ##################### FINAL ANSWER #####################")
	fmt.Println(" #########################################################")
	fmt.Println(`             points_small: `, finalAnswerPointsSmall)
	fmt.Println(`             points_large: `, finalAnswerPointsLarge)
	fmt.Println(" #########################################################")
	fmt.Println(" #########################################################")
}

// Creates a new instance of the Data struct
func CreateNewDataSet() (*Data, error) {
	d := Data{}
	return &d, nil
}

// Finds sets of 4 points that sum to 100 and calculates the volume of the tetrahedron
// if the volume is smaller than the current smallest tetrahedron, it replaces d.smallestTetra with the new tetrahedron
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

// Parse patterns from the file that match the floatPattern regex
// save the first 3 floats as the location of the point
// save the 4th int as the nVal of the point
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

// I rewrote the findVolume function provided from python to go
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

// a test function to verify the findVolume function is working properly
func testFindVolume() {
	// # Example points
	A := []float64{1, 2, 3}
	B := []float64{2, 3, 4}
	C := []float64{1, 5, 1}
	D := []float64{4, 2, 3}

	vol := findVolume(A, B, C, D)
	fmt.Println("The volume of the tetrahedron is ", vol)
}
