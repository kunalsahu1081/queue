
package main

import (
	"bufio"
	"fmt"
	"os"
)

var (
	in  = bufio.NewReader(os.Stdin)
	out = bufio.NewWriter(os.Stdout)
)

func nextInt() int {
	var n int
	fmt.Fscan(in, &n)
	return n
}

func nextString() string {
	var s string
	fmt.Fscan(in, &s)
	return s
}

func solve() {



		
		
	n, k := nextInt(), nextInt()
	
	arr := []int{}

	for i := 0; i < n; i++ {
		num := nextInt()
		arr = append(arr, num)
	}

	ans := 0

	for i := 0; i < n; i++ {

		k0 := min(i, k);

		d2c := 0

		if k0 > 0 {
			if k0 % 2 == 0 {
				d2c++
			}

			if k0 % 4 == 0 {
				d2c++
			}

			if (k0 + 1) % 2 == 0 {
				d2c++
			}

			if (k0 + 1) % 4 == 0 {
				d2c++
			}
		}

		dist := (i - k0) * k

		c2 := 0

		if dist >0 && dist % 2 == 0 {
			c2++
		}

		if arr[i] % 2 == 0 {
			continue
		}

		if k0 == 0 {

			
			ans = (ans + 1) % 2
			

		} else {

			if dist <= 0 {

				if d2c >= 2 {
					ans = (ans + 1) % 2
				}

			} else  {

				if d2c >= 2 && dist % 2 == 1 {
				} else if d2c < 2 && dist %2 == 0 {
					
				} else {
					ans = (ans + 1)% 2
				}

			}

		}



	}


	if ans == 0 {
		fmt.Fprintln(out, "Bob")
	} else {
		fmt.Fprintln(out, "Alice")
	}
	
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func main() {
	defer out.Flush()

	t := nextInt() // Number of test cases

	for ; t > 0; t-- {
		solve()
	}
}
