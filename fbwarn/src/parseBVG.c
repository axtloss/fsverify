#include <stdbool.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "BVGTypes.h"
#include "raylib.h"
#include "extString.h"

void BVGDrawRectangle(BVGRectangle *rectangle) {
  if (rectangle->fill) {
    DrawRectangleRec(rectangle->rayrectangle, rectangle->color);
  } else {
    DrawRectangleLinesEx(rectangle->rayrectangle, rectangle->lineThickness, rectangle->color);
  }
  return;
}

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

/*
  Converts a given char pointer to a corresponding boolean value
  expects the char to only contain true/false without any whitespace or similiar
  case is ignored
  If no match is found, true is returned by default
*/
bool parseBoolValue(char *value) {
  char *valueLower = strdup(strlwr(value));
  printf("Matching bool %s --- %s", valueLower, value);
  if (strcmp(valueLower, "false") == 0) {
    return false;
  } else {
    return true;
  }
}

BVGRectangle *BVGParseRectangle(char *argv[7]) {
  BVGRectangle *result = malloc(sizeof(BVGRectangle));
  Rectangle *rectangle = malloc(sizeof(Rectangle));
  argv[4][strlen(argv[4])-1] = '\0';
  argv[4] = argv[4]+strlen("color=#'");
  int x, y, width, height, r, g, b, a;
  float thickness = 1.0;
  Color *clr = malloc(sizeof(Color));
  sscanf(argv[0]+strlen("x="), "%d", &x);
  sscanf(argv[1]+strlen("y="), "%d", &y);
  sscanf(argv[2]+strlen("width="), "%d", &width);
  sscanf(argv[3]+strlen("height="), "%d", &height);
  if (strlen(argv[4]) == 6) {
    sscanf(argv[4], "%02x%02x%02x", &r, &g, &b);
    clr->r=r; clr->g=g; clr->b=b; clr->a=255;
  } else {
    sscanf(argv[4], "%02x%02x%02x%02x", &r, &g, &b, &a);
    clr->r=r; clr->g=g; clr->b=b; clr->a=a;
  }
  sscanf(argv[6]+strlen("thickness="), "%fd", &thickness);
  printf("X: %d, Y: %d\n", x, y);
  printf("Width: %d, Height: %d\n", width, height);
  printf("Color: %d, %d, %d\n", r, g, b);
  printf("Fill: %d, Thickness: %f\n", parseBoolValue(argv[5]+strlen("fill=")), thickness);

  rectangle->x=x; rectangle->y=y;
  rectangle->height=height; rectangle->width=width;
  result->rayrectangle=*rectangle;

  result->color=*clr;
  result->fill=parseBoolValue(argv[5]+strlen("fill="));
  result->lineThickness=thickness;
  
  return result;
}

void matchFunctionCall(char *call, void *ret) {
  printf("Matching %s\n", call);
  char *funcCall = strdup(call);
  char *funcName = strsep(&funcCall, "(");
  printf("Got function %s\n", funcName);
  if (strcmp(funcName, "rectangle ") == 0) {
      char **arg, *argv[7];
      call = call+strlen("rectangle (");
      for (arg = argv; (*arg = strsep(&call, ",")) != NULL;)
	if (**arg != '\0')
	  if (++arg >= &argv[7])
	    break;
      BVGRectangle *rect = malloc(sizeof(BVGRectangle)); 
      rect = BVGParseRectangle(argv);
      BVGDrawRectangle(rect);
  }
  return;
}
