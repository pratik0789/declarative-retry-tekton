BEGIN { FS=","; OFS=","; print "size,position,trials,meanBaselineMs,meanRecoveryMs,meanTasksAvoided,meanTaskAvoidancePercent,meanTimeReductionPercent" }
NR == 1 { next }
{
  key=$1 SUBSEP $2
  count[key]++
  baseline[key]+=$5
  recovery[key]+=$6
  avoided[key]+=$9
  task_pct[key]+=$10
  time_pct[key]+=$11
}
END {
  for (key in count) {
    split(key, parts, SUBSEP)
    printf "%s,%s,%d,%.2f,%.2f,%.2f,%.2f,%.2f\n", parts[1], parts[2], count[key], baseline[key]/count[key], recovery[key]/count[key], avoided[key]/count[key], task_pct[key]/count[key], time_pct[key]/count[key]
  }
}
