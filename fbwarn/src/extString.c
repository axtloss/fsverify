#include <stdlib.h>
#include <string.h>
#include <ctype.h>

char *strlwr(char *s)
{
  unsigned char *p = (unsigned char *)s;

  while (*p) {
     *p = tolower((unsigned char)*p);
      p++;
  }

  return s;
}

char *trim(char *s)
{
  char *result = strdup(s);
  char *end;

  while(isspace((unsigned char)*result)) result++;

  if(*result == 0)
    return result;

  end = result + strlen(result) - 1;
  while(end > result && isspace((unsigned char)*end)) end--;

  end[1] = '\0';

  return result;
}

char *replaceStr(char *s, char *old, char *replace) {
  char* result; 
  int i, cnt = 0; 
  size_t newSize = strlen(replace); 
  size_t oldSize = strlen(old); 
 
  for (i = 0; s[i] != '\0'; i++) { 
    if (strstr(&s[i], old) == &s[i]) { 
      cnt++; 
      i += oldSize - 1; 
    } 
  } 

  result = (char*)malloc(i + cnt * (newSize - oldSize) + 1); 
 
  i = 0; 
  while (*s) { 
    if (strstr(s, old) == s) { 
      strcpy(&result[i], replace); 
      i += newSize; 
      s += oldSize; 
    } 
    else
      result[i++] = *s++; 
  } 
 
  result[i] = '\0'; 
  return result; 
};
