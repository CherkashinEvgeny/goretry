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
Breaking retry loop:
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
Cancellation:
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