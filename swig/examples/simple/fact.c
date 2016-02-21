#include "fact.h"
#include "limits.h"
#include "stdio.h"

long fact(long n) {
  printf("newmath.fact> %ld\n", n);
  // termination of recursion
  if( n <= 1 ) return 1;

  if( INT_MAX/n < n-1 ) {
    printf("newmath.fact> overflow %ld\n", n);
    return 1;
  }
  return n*fact(n-1);
}

