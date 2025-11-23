package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Resoult struct {
	nr      int
	mainNum [5]int
	subNum  [2]int
}

func getResuolts(ctx context.Context) []Resoult {
	today := time.Now()
	url := fmt.Sprintf("https://megalotto.pl/wyniki/eurojackpot/losowania-od-3-Stycznia-2017-do-%d-%d-%d", today.Day(), today.Month(), today.Year())

	res, err := http.Get(url)
	if err != nil {
		ctx.Err()
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		ctx.Err()
	}

	var resoults []Resoult

	doc.Find("div.lista_ostatnich_losowan > ul").Each(func(i int, s *goquery.Selection) {
		var nr int
		integer, err := strconv.ParseInt(strings.Split(s.Find("li.nr_in_list").Text(), ".")[0], 10, 32)
		if err != nil {
			ctx.Err()
		}
		nr = int(integer)

		var mainNum [5]int
		s.Find("li.numbers_in_list").Each(func(i int, s *goquery.Selection) {
			integer, err := strconv.ParseInt(strings.Trim(s.Text(), " "), 10, 32)
			if err != nil {
				ctx.Err()
			}
			mainNum[i] = int(integer)
		})

		var subNum [2]int
		s.Find("li.tsn_number_in_list").Each(func(i int, s *goquery.Selection) {
			integer, err := strconv.ParseInt(strings.Trim(s.Text(), " \n"), 10, 32)
			if err != nil {
				ctx.Err()
			}
			subNum[i] = int(integer)
		})

		resoults = append(resoults, Resoult{nr, mainNum, subNum})

	})
	return resoults
}

type NumsType struct {
	num   int
	count int
}

type Statistics struct {
	mainNums [50]NumsType
	subNums  [12]NumsType
}

func getStatistics(resoults []Resoult) Statistics {
	stats := Statistics{mainNums: [50]NumsType{}, subNums: [12]NumsType{}}

	for i := 0; i < 50; i++ {
		stats.mainNums[i] = NumsType{num: i + 1, count: 0}
	}
	for i := 0; i < 12; i++ {
		stats.subNums[i] = NumsType{num: i + 1, count: 0}
	}

	for _, r := range resoults {
		for i, mN := range stats.mainNums {
			for _, rN := range r.mainNum {
				if mN.num == rN {
					stats.mainNums[i].count++
				}
			}
		}

		for i, sN := range stats.subNums {
			for _, rN := range r.subNum {
				if sN.num == rN {
					stats.subNums[i].count++
				}

			}
		}
	}

	return stats
}

func printStats(stats Statistics) {
	fmt.Printf("mainNums: \n")
	for _, mN := range stats.mainNums {
		fmt.Printf("%+3v\n", mN)
	}
	fmt.Printf("\nsubNums: \n")
	for _, sN := range stats.subNums {
		fmt.Printf("%+4v\n", sN)
	}
}

func getRandomNumbers() ([5]int, [2]int) {
	randMainNum := [5]int{}
	randSubNum := [2]int{}

	for i := 0; i < 5; i++ {
		r := rand.Intn(50-1) + 1
		for slices.Contains(randMainNum[:], r) {
			r = rand.Intn(50-1) + 1
		}
		randMainNum[i] = r
	}

	for i := 0; i < 2; i++ {
		r := rand.Intn(12-1) + 1
		for slices.Contains(randSubNum[:], r) {
			r = rand.Intn(12-1) + 1
		}
		randSubNum[i] = r
	}

	return randMainNum, randSubNum
}

func isRandMainNumOri(randMain [5]int, resoults []Resoult) bool {
	for _, r := range resoults {
		set := slices.Concat(randMain[:], r.mainNum[:])
		slices.Sort(set)
		set = slices.Compact(set)
		if len(set) < 8 {
			return false
		}
	}
	return true
}

func checkRandNum(ranMainNum [5]int, ranSubNum [2]int, resoults []Resoult, stats Statistics) bool {
	return isRandMainNumOri(ranMainNum, resoults)
}

func run(ctx context.Context, w io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)

	defer cancel()

	res := getResuolts(ctx)

	stats := getStatistics(res)

	printStats(stats)

	randMain, randSub := getRandomNumbers()

	fmt.Println("\n\n", "first:\n", randMain, randSub)

	isOK := checkRandNum(randMain, randSub, res, stats)

	for !isOK {
		randMain, randSub = getRandomNumbers()

		fmt.Println("\n\n", "repeated:\n", randMain, randSub)

		isOK = checkRandNum(randMain, randSub, res, stats)
	}

	slices.Sort(randMain[:])
	slices.Sort(randSub[:])

	fmt.Println("\n\n", "final: \n", randMain, randSub)

	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s/n", err)
		os.Exit(1)
	}
}
