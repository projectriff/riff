# Gnuplot script: http://www.gnuplot.info/

set terminal png truecolor size 3000,2000 font 'Droid Sans Mono' 20 enhanced


set ylabel 'Writes and Replicas'

set ytics nomirror
set y2tics
set y2label 'Queue length'

set multiplot layout 1, 1

  set title 'Step Scenario'
  plot "step-scenario.dat" using 1:4  with lines  title "writes"        linecolor "#000000"  axes x1y1, \
                        "" using 1:3  with lines  title "queue length"  linecolor "#bdd4f9"  axes x1y2, \
                        "" using 1:2  with lines  title "replicas"      linecolor "#870202"  axes x1y1

unset multiplot