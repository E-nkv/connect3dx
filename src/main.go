package main

import "fmt"

func printArr(arr [][][]int) {
	for r := range arr {
		for c := range arr[r] {
			fmt.Println(arr[r][c])
		}
	}
}
func main() {
	R, C, H := 4, 4, 4
	arr := make([][][]int, R)
	for i := range R {
		arr[i] = make([][]int, C)
		for j := range C {
			arr[i][j] = make([]int, H)
		}
	}
	arr[3][0][0] = 1
	arr[3][1][0] = 1
	arr[3][2][0] = 1
	arr[3][3][0] = 2
	arr[2][3][0] = 1
	printArr(arr)

	/* app := api.NewApp()
	m := http.NewServeMux()
	m.HandleFunc("/ws", app.HandleWs)
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("hi")
		fmt.Fprint(w, "hi")
	})
	fmt.Println("running at 8080")
	log.Fatal(http.ListenAndServe(":8080", m)) */
}
