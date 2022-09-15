#!/bin/bash
set -e
STACK="cflinuxfs3"
runbuildback() {
    echo "You requested stack: $STACK"
    buildpack-packager build -stack $STACK --cached true
}

runarray() {
    cd /build
    ls -la
    for d in */ ; do
      echo "Checking $d..."
      if [ -f "$d/build.sh" ]; then
           if [ ! -f "$d/completed" ]; then
             source "$d/build.sh"
             STACK=`cat /build/$d/stack`
             runbuildback
             touch "/build/$d/completed"
             else
                echo "Already Completed"
           fi
      fi
    done
}

runarray
