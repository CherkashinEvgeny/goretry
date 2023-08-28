# retry

Golang retry library.

## About The Project

Retry - small library, that provides api for easy function retry and error handling.
Package has some benefits unlike known implementations:

- laconic api
- rich retry strategy set
- easy customization
- context package support

## Usage

Retry function until success:

```
_ = retry.Exec(func(retryNumber int) (err error) {
    // your logic here
    return
})
```

Limit attempts:

```
err := retry.Exec(func(retryNumber int) (err error) {
    // your logic here
    return
}, retry.MaxAttempts(10))
```

Delay between retries:

```
err := retry.Exec(func(retryNumber int) (err error) {
    // your logic here
    return
}, retry.FixedDelay(time.Second))
```

Combine few retry strategies:

```
err := retry.Exec(func(retryNumber int) (err error) {
    // your logic here
    return
}, retry.MaxAttempts(10), retry.FixedDelay(time.Second))
```

Breaking retry loop directly from function:

```
err := retry.Exec(func(retryNumber int) (err error) {
    // your logic here
    if !canRetry {
        err = retry.Unrecoverable(err)
        return
    }
    return
})
```

Using context:

```
err := retry.ExecContext(ctx, func(ctx context.Context, retryNumber int) (err error) {
    // your logic here
    return
})
```

Context cancellation:

```
ctx, cancel := context.WithCancel(context.Background())
go func() {
    // wait some event
    cancel()
}()
err := retry.ExecContext(ctx, func(ctx context.Context, retryNumber int) (err error) {
    // your logic here
    return
})
```

## Underwater rocks

### Arguments order

Be careful:

```
err := retry.Exec(func(retryNumber int) (err error) {
    return errors.New("test")
}, retry.MaxAttempts(2), retry.FixedDelay(time.Second))
```

not similar to

```
err := retry.Exec(func(retryNumber int) (err error) {
    return errors.New("test")
}, retry.FixedDelay(time.Second), retry.MaxAttempts(2))
```

Trace steps in the first case are:

1. Function call
2. Max attempts check
3. Delay
4. Function call
5. Max attempts check
6. Break loop

Trace steps in the second case are:

1. Function call
2. Delay
3. Max attempts check
4. Function call
5. Delay
6. Max attempts check
7. Break loop

So, we can see, that second block of code will have unnecessary delay.
It is highly recommended to store strategies, that limit attempts, before strategies, that do delays.

### Strategy reusing

Some strategies have internal state, so it is highly recommended not to share strategies between several `Exec` calls.

Incorrect:

```
strategy := retry.MaxAttempts(2)
err := retry.Exec(func(retryNumber int) (err error) {
    // your logic here
    return
}, strategy)
err := retry.Exec(func(retryNumber int) (err error) {
    // your logic here
    return
}, strategy)
```

Correct:

```
err := retry.Exec(func(retryNumber int) (err error) {
    return
}, retry.MaxAttempts(2))
err := retry.Exec(func(retryNumber int) (err error) {
    return
}, retry.MaxAttempts(2))
```

## Similar projects

- [avast/retry-go](https://github.com/avast/retry-go)
- [Rican7/retry](https://github.com/Rican7/retry)
- [kamilsk/retry](https://github.com/kamilsk/retry)

## License

Retry is licensed under the Apache License, Version 2.0. See [LICENSE](./LICENCE.md)
for the full license text.

## Contact

- Email: `cherkashin.evgeny.viktorovich@gmail.com`
- Telegram: `@evgeny_cherkashin`