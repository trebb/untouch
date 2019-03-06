#!/bin/sh

RTC=/dev/rtc0

[ ! -c $RTC ] && exit 0

[ ! -x /sbin/hwclock ] && exit 0

[ "$UTC" = "yes" ] && tz="--utc" || tz="--localtime"

case "$1" in
    start)
        echo "Restoring system time from RTC"
        if [ -z "$TZ" ]; then
            hwclock $tz --hctosys -f $RTC
        else
            TZ="$TZ" hwclock $tz --hctosys -f $RTC
        fi
        ;;

    stop)
        echo "Saving system time to RTC"
        hwclock $tz --systohc -f $RTC
        ;;

    show)
        hwclock $tz --show -f $RTC
        ;;

    *)
        echo "Usage: hwclock.sh {start|stop|show}" >&2
        exit 1
esac
