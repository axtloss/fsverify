#include <stdlib.h>
#include <string.h>
#include <ctype.h>

char *strlwr(char *str)
{
  unsigned char *p = (unsigned char *)str;

  while (*p) {
     *p = tolower((unsigned char)*p);
      p++;
  }

  return str;
}

char *trim(char *str)
{
  char *result = strdup(str);
  char *end;

  while(isspace((unsigned char)*result)) result++;

  if(*result == 0)
    return result;

  end = result + strlen(result) - 1;
  while(end > result && isspace((unsigned char)*end)) end--;

  // Write new null terminator character
  end[1] = '\0';

  return result;
}
