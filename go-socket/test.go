package main
import (
	"fmt"
	"math/big"
)

type BlockObject struct {
	Blocks []*big.Int `json:"blocks"`
	MaxNumber *big.Int `json:"max"`
}

func calculateDifferences(blockObjects []BlockObject) []*big.Int {
	differences := make([]*big.Int, 0)
	diffs_sequence := big.NewInt(0)
	diffs_alchemy := big.NewInt(0)
	diffs_quicknode := big.NewInt(0)
	diffs_polygon := big.NewInt(0)
	diffs_ankr := big.NewInt(0)

	for _, blockObj := range blockObjects {
		var diffs []*big.Int
		diffs_sequence.Add(diffs_sequence, blockObj.Blocks[0])
		diffs_alchemy.Add(diffs_alchemy, blockObj.Blocks[1])
		diffs_quicknode.Add(diffs_quicknode, blockObj.Blocks[2])
		diffs_polygon.Add(diffs_polygon, blockObj.Blocks[3])
		diffs_ankr.Add(diffs_ankr, blockObj.Blocks[4])
		diffs = append(diffs, diffs_sequence)
		diffs = append(diffs, diffs_alchemy)
		diffs = append(diffs, diffs_quicknode)
		diffs = append(diffs, diffs_polygon)
		diffs = append(diffs, diffs_ankr)
		differences = diffs
	}

	return differences
}

func main() {

	var array []*big.Int

	array = append(array, big.NewInt(4))
	array = append(array, big.NewInt(7))
	array = append(array, big.NewInt(2))
	array = append(array, big.NewInt(9))
	array = append(array, big.NewInt(5))


	blockObjects := []BlockObject{
		{
			Blocks:    array,
			MaxNumber: big.NewInt(4),
		},
		{
			Blocks:    array,
			MaxNumber: big.NewInt(3),
		},
	}

	differences := calculateDifferences(blockObjects)

	fmt.Println("Differences:")
	for _, diff := range differences {
		fmt.Println(diff)
	}
}
