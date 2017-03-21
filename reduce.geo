let reduce = fn(f, seed, arr) {
	let iter = fn(acc, arr) {
		if (len(arr) == 0) {
			return acc
		}
		iter(f(acc, head(arr)), tail(arr))
	};

	iter(seed, arr);
};

let sum = fn(arr) {
	reduce(fn(acc, it) { acc + it }, 0, arr);
};

[1, 2, 3, 4, 5] | sum;
