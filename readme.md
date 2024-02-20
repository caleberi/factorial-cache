### Factorial Cache
A simple benchmark test of how datatype size affect during some important calculation such as factorial. In this test, I used an http server to test there reponse speed of factorial calculation of a given number n provided by a client.

### Assumption
there is a need to precompute the value provided by client and store it on disk to better understand how and when to flush information to disk and also how to load such information in memory on startup .

### Benchmarks and Testing
Initial testing with maximum number n = 500  shows the following result

- ![ factorial memoized with big interger](./Screenshot%202024-02-20%20at%2013.52.42.png)
- ![factorial memoized with interger](./Screenshot%202024-02-20%20at%2013.55.20.png)
- ![factorial with no memoization](./Screenshot%202024-02-20%20at%2013.56.47.png)

Benchmarking showed significant use of memory for factorial with no memoization and factorial memoized with interger

However the test was measured in terms of overflow and failure (this involves factorial with zero result and negative result due to overflow)
