# Gnuplot script: http://www.gnuplot.info/

set terminal png truecolor size 3000,2000 font 'Droid Sans Mono' 20 enhanced


set ylabel 'Writes and Replicas'

set ytics nomirror
set y2tics
set y2label 'Queue length'

WRITES_STYLE  ="filledcurve fs solid 0.5 noborder linecolor '#efecbf'"
QUEUES_STYLE  ="lines linecolor '#bdd4f9'"
REPLICAS_STYLE="lines linecolor '#870202'"

set multiplot layout 3, 1
  set title 'Step Scenario'
  set logscale y
  plot "step-scenario.dat" using 1:4  title "writes"       with @WRITES_STYLE   axes x1y1, \
                        "" using 1:3  title "queue length" with @QUEUES_STYLE  axes x1y2, \
                        "" using 1:2  title "replicas"     with @REPLICAS_STYLE axes x1y1
  unset logscale y

  set title 'Sinusoidal Scenario'
  plot "sine-scenario.dat" using 1:4  title "writes"       with @WRITES_STYLE   axes x1y1, \
                        "" using 1:3  title "queue length" with @QUEUES_STYLE  axes x1y2, \
                        "" using 1:2  title "replicas"     with @REPLICAS_STYLE axes x1y1

  set title 'Ramp Scenario'
  plot "ramp-scenario.dat" using 1:4  title "writes"       with @WRITES_STYLE   axes x1y1, \
                        "" using 1:3  title "queue length" with @QUEUES_STYLE  axes x1y2, \
                        "" using 1:2  title "replicas"     with @REPLICAS_STYLE axes x1y1
unset multiplot