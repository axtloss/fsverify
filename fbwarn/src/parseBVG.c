#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "raylib.h"

// reads a given file
FILE *readFile(char *path) {
  FILE *fp = fopen (path, "r");
  if (fp == NULL) {
    perror ("Failed to open file!");
    exit(1);
  }
  return fp;
}

/*
  Converts a multiline function
  call into a single line call
*/
char *multiToSingle(char *lines) {
  // allocating the size of lines is safe since characters arent added
  char *result = malloc (strlen(lines)+1);
  char *line;
  while ((line = strsep(&lines, "\n")) != NULL)
    sprintf(result, "%s%s", result, line);

  free(line);
  return result; 
}

void BVGRectangle(char *argv[5]) {
  printf("Drawing rectangle\n");
  argv[4][strlen(argv[4])-2] = '\0';
  printf("%s, %s\n %s, %s,\n %s\n", argv[0], argv[1], argv[2], argv[3], argv[4]);
  int x, y, width, height, r, g, b;
  sscanf(argv[0]+strlen("x="), "%d", &x);
  sscanf(argv[1]+strlen("y="), "%d", &y);
  sscanf(argv[2]+strlen("width="), "%d", &width);
  sscanf(argv[3]+strlen("height="), "%d", &height);
  sscanf(argv[4]+strlen("color='#"), "%02x%02x%02x", &r, &g, &b);
  printf("X: %d, Y: %d\n", x, y);
  printf("Width: %d, Height: %d\n", width, height);
  printf("Color: %d, %d, %d\n", r, g, b);
  Color *clr = malloc(sizeof(Color));
  clr->r=r;
  clr->g=g;
  clr->b=b;
  clr->a=255;
  DrawRectangle(x, y, width, height, *clr);
  return;
}

void matchFunctionCall(char *call) {
  printf("Matching %s\n", call);
  char *funcCall = strdup(call);
  char *funcName = strsep(&funcCall, "(");
  printf("Got function %s\n", funcName);
  if (strcmp(funcName, "rectangle ") == 0) {
      char **arg, *argv[5];
      call = call+strlen("rectangle (");
      for (arg = argv; (*arg = strsep(&call, ",")) != NULL;)
	if (**arg != '\0')
	  if (++arg >= &argv[5])
	    break;
      BVGRectangle(argv);
  }
  return;
}
