## Runtime {#runtime}

<!-- go.dev/issue/36141 -->
On Windows, the monotonic clock used by [time.Now], [time.Since], [time.Sub],
and the runtime's timers and tickers now measures "program time", which stops
while the system is asleep, matching the behavior on other operating systems.
Previously it continued to advance while the system was asleep. The previous
behavior can be restored by setting `GODEBUG=wintime=external`. Programs that
need to measure elapsed real time across sleep should use the new
`time.External*` functions instead.
