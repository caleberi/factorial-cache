package main

import (
	"bufio"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var testMode = false

type tools struct {
}

type Application struct {
	fact_lock      sync.Mutex
	fib_lock       sync.Mutex
	big_fact_lock  sync.Mutex
	fact_cache     map[int]int
	fib_cache      map[int]int
	big_fact_cache map[*big.Int]big.Int
	server         *http.ServeMux
	quit           chan os.Signal
}

func NewApp() *Application {
	return &Application{
		fact_lock:      sync.Mutex{},
		fib_lock:       sync.Mutex{},
		big_fact_lock:  sync.Mutex{},
		fact_cache:     make(map[int]int, 1),
		fib_cache:      make(map[int]int, 1),
		big_fact_cache: make(map[*big.Int]big.Int),
	}
}

func (app *Application) serve() *Application {
	app.quit = make(chan os.Signal, 1)
	signal.Notify(app.quit, syscall.SIGINT, syscall.SIGTERM)

	var err error
	app.fib_cache, err = loadIntCache("./fib-history.txt")
	if err != nil {
		log.Fatalln(err)
	}

	app.fact_cache, err = loadIntCache("./fact-history.txt")
	if err != nil {
		log.Fatalln(err)
	}

	app.big_fact_cache, err = loadStringCache("./fact-big-history.txt")
	if err != nil {
		log.Fatalln(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/factorial", app.factorialHandler)
	mux.HandleFunc("/fibonacci", app.fibonacciHandler)
	mux.HandleFunc("/factorial-big", app.factorialBigHandler)
	mux.HandleFunc("/fibonacci-no-memo", app.fibonacciNonHandler)
	mux.HandleFunc("/factorial-no-memo", app.factorialNonHandler)

	app.server = mux

	go func(quit chan os.Signal, fibCache, factCache map[int]int, factBigCache map[*big.Int]big.Int) {
		<-quit //  wait till we are signal that we are done
		// save the data into the file
		wg := sync.WaitGroup{}
		wg.Add(3)
		// handle cleanup when done with factorial server
		// write to disk the cache details
		go saveIntCache(&wg, "./fib-history.txt", fibCache)
		go saveIntCache(&wg, "./fact-history.txt", factCache)
		go saveBigIntCache(&wg, "./fact-big-history.txt", factBigCache)

		wg.Wait()

		time.Sleep(1 * time.Second)
		if testMode {
			return
		}
		os.Exit(0)
	}(app.quit, app.fib_cache, app.fact_cache, app.big_fact_cache)

	return app
}

func (app *Application) factorialNoMemo(n int) int {
	if n == 1 || n == 0 {
		return n
	}
	return n * app.factorialNoMemo(n-1)
}

func (app *Application) fibonacciNoMemo(n int) int {
	if n == 1 || n == 0 {
		return n
	}
	return app.fibonacciNoMemo(n-1) + app.fibonacciNoMemo(n-2)
}

func (app *Application) factorialBig(n *big.Int) *big.Int {
	if n.String() == "0" || n.String() == "1" {
		return n
	}
	if val, ok := app.big_fact_cache[n]; ok {
		return &val
	}

	var result big.Int
	result.Sub(n, big.NewInt(1)).Mul(n, app.factorialBig(&result)) //
	app.big_fact_lock.Lock()
	app.big_fact_cache[n] = result
	app.big_fact_lock.Unlock()
	return &result
}
func (app *Application) factorial(n int) int {
	if n == 0 || n == 1 {
		return 1
	}
	if val, ok := app.fact_cache[n]; ok {
		return val
	}

	result := n * app.factorial(n-1) //
	app.fact_lock.Lock()
	app.fact_cache[n] = result
	app.fact_lock.Unlock()
	return result
}

func (app *Application) fibonacci(n int) int {
	if n == 1 || n == 0 {
		return n
	}

	if val, ok := app.fib_cache[n]; ok {
		return val
	}

	result := app.fibonacci(n-1) + app.fibonacci(n-2)
	app.fib_lock.Lock()
	app.fib_cache[n] = result
	app.fib_lock.Unlock()
	return app.fib_cache[n]
}

func (app *Application) factorialHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	numStr := query.Get("n")
	num, ok := extraNum(numStr, w)
	if !ok {
		fmt.Printf("error changing num :  %s\n", numStr)
		return
	}
	result := app.factorial(num)
	fmt.Fprintf(w, "%d\n", result)
}

func extraNum(numStr string, w http.ResponseWriter) (int, bool) {
	if numStr == "" {
		http.Error(w, "Parameter 'n' is missing", http.StatusBadRequest)
		return 0, false
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		http.Error(w, "Invalid parameter 'n'", http.StatusBadRequest)
		return 0, false
	}

	if num < 0 {
		http.Error(w, "Invalid parameter 'n': factorial is not defined for negative numbers", http.StatusBadRequest)
		return 0, false
	}

	return num, true
}

func (app *Application) factorialNonHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	numStr := query.Get("n")
	num, ok := extraNum(numStr, w)
	if !ok {
		fmt.Printf("error changing num :  %s\n", numStr)
		return
	}

	result := app.factorialNoMemo(num)

	fmt.Fprintf(w, "%d\n", result)
}

func (app *Application) fibonacciNonHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	numStr := query.Get("n")
	num, ok := extraNum(numStr, w)
	if !ok {
		fmt.Printf("error changing num :  %s\n", numStr)
		return
	}

	result := app.fibonacciNoMemo(num)

	fmt.Fprintf(w, "%d\n", result)
}

func (app *Application) fibonacciHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	numStr := query.Get("n")
	num, ok := extraNum(numStr, w)
	if !ok {
		fmt.Printf("error changing num :  %s\n", numStr)
		return
	}

	result := app.fibonacci(num)

	fmt.Fprintf(w, "%d\n", result)
}

func (app *Application) factorialBigHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	numStr := query.Get("n")
	if numStr == "" || strings.HasPrefix(numStr, "-") {
		numStr = "0"
	}
	bn := new(big.Int)
	bn.SetString(numStr, 10)
	result := app.factorialBig(bn)

	fmt.Fprintf(w, "%s", result.String())
}

func main() {
	app := NewApp()
	fmt.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", app.serve().server); err != nil {
		log.Fatalln(err)
	}
}

func loadIntCache(filename string) (map[int]int, error) {
	values := make(map[int]int)

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Invalid line format: %s\n", line)
			continue
		}
		key := strings.TrimSpace(parts[0])
		ikey, err := strconv.Atoi(key)
		if err != nil {
			fmt.Printf("Invalid key format: %s\n", key)
			continue
		}
		value := strings.TrimSpace(parts[1])
		ivalue, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("Invalid key format: %s\n", value)
			continue
		}

		values[ikey] = ivalue
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func loadStringCache(filename string) (map[*big.Int]big.Int, error) {
	values := make(map[*big.Int]big.Int)

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Invalid line format: %s\n", line)
			continue
		}
		key := big.NewInt(0)
		key.SetString(strings.TrimSpace(parts[0]), 10)
		value := big.NewInt(0)
		value.SetString(strings.TrimSpace(parts[1]), 10)
		values[key] = *value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func saveIntCache(wg *sync.WaitGroup, filename string, data map[int]int) {
	defer wg.Done()
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	err = os.Truncate(filename, 0)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		log.Fatalln(err)
	}

	writer := bufio.NewWriter(file)
	for key, value := range data {
		_, err := fmt.Fprintf(writer, "%d=%d\n", key, value)
		if err != nil {
			log.Println(err)
			return
		}
	}
	err = writer.Flush()
	if err != nil {
		log.Println(err)
		return
	}
}

func saveBigIntCache(wg *sync.WaitGroup, filename string, data map[*big.Int]big.Int) {
	defer wg.Done()
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	err = os.Truncate(filename, 0)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		log.Fatalln(err)
	}

	writer := bufio.NewWriter(file)
	for key, value := range data {
		_, err := fmt.Fprintf(writer, "%s=%s\n", key.String(), value.String())
		if err != nil {
			log.Println(err)
			return
		}
	}
	err = writer.Flush()
	if err != nil {
		log.Println(err)
		return
	}
}
