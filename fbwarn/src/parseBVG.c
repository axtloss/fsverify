#include <stdbool.h>
#include <stddef.h>
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

void BVGDrawRoundedRectangle(BVGRoundedRectangle *rectangle) {
  if (rectangle->rectangle.fill) {
    DrawRectangleRounded(rectangle->rectangle.rayrectangle, rectangle->roundness, rectangle->segments, rectangle->rectangle.color);
  } else {
    DrawRectangleRoundedLines(rectangle->rectangle.rayrectangle, rectangle->roundness, rectangle->segments, rectangle->rectangle.lineThickness, rectangle->rectangle.color);
  }
}

void BVGDrawCircle(BVGCircle *circle) {
  Vector2 center = {circle->centerX, circle->centerY};
  if (circle->drawSector) {
    DrawCircleSector(center, circle->radius, circle->startAngle, circle->endAngle, circle->segments, circle->color);
  } else {
    printf("center: %f, %f\n", center.x, center.y);
    printf("radius: %f, color: %d\n", circle->radius, circle->color.a);
    DrawCircle(circle->centerX, circle->centerY, circle->radius, circle->color);
  }
}

void BVGDrawRing(BVGRing *ring) {
  Vector2 center = {ring->centerX, ring->centerY};
  DrawRing(center, ring->inRadius, ring->outRadius, ring->startAngle, ring->endAngle, ring->segmets, ring->color);
}

void BVGDrawText(BVGText *text) {
  DrawText(text->text, text->x, text->y, text->fontSize, text->color);
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
char *multiToSingle(char *s) {
  // allocating the size of lines is safe since characters arent added
  char *midresult = malloc (strlen(s)+1);
  char *line;
  midresult = replaceStr(s, "\n", "");
  return replaceStr(midresult, ", ", ",");
}

/*
  Collects all arguments from a function call
  arguments are seperated with a ,
*/
void collectArgs(char *res[], char *call, int n) {
  char **arg;
  char *args[n];
  for (arg = args; (*arg = strsep(&call, ",")) != NULL;)
	if (**arg != '\0')
	  if (++arg >= &args[n])
 	    break;
  memcpy(res, args, sizeof(args));
}

/*
  Order the arguments in the order specified with knownArgs
  the res array contains only the values of each argument
*/
void orderArgs(char *res[], char *argv[], int n, char *knownArgs[]) {
  for(int i=0; i<n; i++) {
    for(int j=0; j<n; j++) {
      if (strncmp(argv[i], knownArgs[j], strlen(knownArgs[j])) == 0) {
	res[j] = argv[i]+strlen(knownArgs[j])+1;
      }
    }
  }
}

/*
  Parses a color from a hex representation
  supports RRGGBB and RRGGBBAA
  assumes that only the color value is in the string
*/
Color *parseColorFromHex(char *hex) {
  Color *clr = malloc(sizeof(Color));
  int r,g,b,a;
  if (strlen(hex) == 6) {
    sscanf(hex, "%02x%02x%02x", &r, &g, &b);
    clr->r=r; clr->g=g; clr->b=b; clr->a=255;
  } else {
    sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &a);
    clr->r=r; clr->g=g; clr->b=b; clr->a=a;
  }
  return clr;
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

/*
  Creates a BVGRectangle based on a BVG function call
*/
BVGRectangle *BVGParseRectangle(char *argv[7]) {
  BVGRectangle *result = malloc(sizeof(BVGRectangle));
  Rectangle *rectangle = malloc(sizeof(Rectangle));
  size_t argN = 7;
  char *args[argN];
  char *knownArgs[7] = {"x", "y", "width", "height", "color", "fill", "thickness"};
  //printf("118 Parsing...%s \n", args[0]);
  orderArgs(args, argv, argN, knownArgs);
  printf("119 Parsing...%s \n", args[0]);
  int x, y, width, height, r, g, b, a;
  float thickness = 1.0;
  bool fill = true;
  sscanf(args[0], "%d", &x);
  sscanf(args[1], "%d", &y);
  sscanf(args[2], "%d", &width);
  sscanf(args[3], "%d", &height);
  args[4] = args[4]+2;
  args[4][strlen(args[4])-1] = '\0';
  Color *clr = parseColorFromHex(args[4]);
  sscanf(args[6], "%fd", &thickness);
  fill = parseBoolValue(args[5]);
  printf("X: %d, Y: %d\n", x, y);
  printf("Width: %d, Height: %d\n", width, height);
  printf("Color: %d, %d, %d\n", r, g, b);
  printf("Fill: %d, Thickness: %f\n", fill, thickness);

  rectangle->x=x; rectangle->y=y;
  rectangle->height=height; rectangle->width=width;
  result->rayrectangle=*rectangle;
  result->color=*clr;
  result->fill=fill;
  result->lineThickness=thickness;
  return result;
}

/*
  Creates a BVGRoundedrectangle based on a BVG function call
*/
BVGRoundedRectangle *BVGParseRoundedRectangle(char *argv[9]) {
  BVGRoundedRectangle *result = malloc(sizeof(BVGRoundedRectangle));
  BVGRectangle *bvgrectangle = malloc(sizeof(BVGRectangle));
  Rectangle *rectangle = malloc(sizeof(Rectangle));
  size_t argN = 9;
  char *args[argN];
  char *knownArgs[9] = {"x", "y", "width", "height", "color", "fill", "thickness", "roundness", "segments"};
  orderArgs(args, argv, argN, knownArgs);
  bvgrectangle = BVGParseRectangle(argv);
  result->rectangle = *bvgrectangle;

  float roundness;
  int segments;
  sscanf(args[7], "%fd", &roundness);
  sscanf(args[8], "%d", &segments);

  printf("Roundness: %fd, Segments: %d\n", roundness, segments);
  result->roundness = roundness;
  result->segments = segments;
  
  return result;
}

BVGCircle *BVGParseCircle(char *argv[4]) {
  BVGCircle *result = malloc(sizeof(BVGCircle));
  size_t argN = 4;
  char *args[argN];
  char *knownArgs[4] = {"x", "y", "radius", "color"};
  orderArgs(args, argv, argN, knownArgs);
  int x, y;
  float radius;
  Color *clr;
  sscanf(args[0], "%d", &x);
  sscanf(args[1], "%d", &y);
  sscanf(args[2], "%fd", &radius);
  args[3] = args[3]+2;
  args[3][strlen(args[3])-1] = '\0';
  clr = parseColorFromHex(args[3]);
  printf("X: %d, Y: %d\n", x, y);
  printf("radius: %f, color: %s\n", radius, args[3]);
  result->drawSector=false;
  result->centerX=x;
  result->centerY=y;
  result->color=*clr;
  result->radius=radius;
  return result;
}

BVGCircle *BVGParseCircleSegment(char *argv[7]) {
  BVGCircle *result = malloc(sizeof(BVGCircle));
  size_t argN = 7;
  char *args[argN];
  char *knownArgs[7] = {"x", "y", "radius", "color", "startangle", "endangle", "segments"};
  orderArgs(args, argv, argN, knownArgs);
  int x, y, segments;
  float radius, startAngle, endAngle;
  Color *clr;
  sscanf(args[0], "%d", &x);
  sscanf(args[1], "%d", &y);
  sscanf(args[2], "%fd", &radius);
  sscanf(args[4], "%fd", &startAngle);
  sscanf(args[5], "%fd", &endAngle);
  sscanf(args[6], "%d", &segments);
  args[3] = args[3]+2;
  args[3][strlen(args[3])-1] = '\0';
  clr = parseColorFromHex(args[3]);
  printf("X: %d, Y: %d\n", x, y);
  printf("radius: %f, color: %s\n", radius, args[3]);
  result->drawSector=true;
  result->centerX=x;
  result->centerY=y;
  result->color=*clr;
  result->radius=radius;
  result->segments=segments;
  result->startAngle=startAngle;
  result->endAngle=endAngle;
  return result;
}

BVGRing *BVGParseRing(char *argv[8]) {
  BVGRing *result = malloc(sizeof(BVGRing));
  size_t argN = 8;
  char *args[argN];
  char *knownArgs[8] = {"x", "y", "innerradius", "outerradius", "startangle", "endangle", "segments", "color"};
  orderArgs(args, argv, argN, knownArgs);
  int x, y, segments;
  float innerRadius, outerRadius, startAngle, endAngle;
  Color *clr;
  sscanf(args[0], "%d", &x);
  sscanf(args[1], "%d", &y);
  sscanf(args[2], "%f", &innerRadius);
  sscanf(args[3], "%f", &outerRadius);
  sscanf(args[4], "%f", &startAngle);
  sscanf(args[5], "%f", &endAngle);
  sscanf(args[6], "%d", &segments);
  args[7] = args[7]+2;
  args[7][strlen(args[7])-1] = '\0';
  clr = parseColorFromHex(args[7]);
  result->centerX=x; result->centerY=y;
  result->inRadius=innerRadius; result->outRadius=outerRadius;
  result->startAngle=startAngle; result->endAngle=endAngle;
  result->segmets=segments;
  result->color=*clr;
  return result;
}

BVGText *BVGParseText(char *argv[5]) {
  BVGText *result = malloc(sizeof(BVGText));
  size_t argN = 5;
  char *args[argN];
  char *knownArgs[5] = {"text", "x", "y", "size", "color"};
  orderArgs(args, argv, argN, knownArgs);
  args[0] = args[0]+1;
  args[0][strlen(args[0])-1] = '\0';
  char *text = args[0];
  int x, y, size;
  Color *clr;
  sscanf(args[1], "%d", &x);
  sscanf(args[2], "%d", &y);
  sscanf(args[3], "%d", &size);
  args[4] = args[4]+2;
  args[4][strlen(args[4])-1] = '\0';
  printf("Text: %s\n", text);
  printf("X: %d, Y: %d\n", x, y);
  printf("Size: %d, Color: %s\n", size, args[4]);
  clr = parseColorFromHex(args[4]);
  result->text=text; result->x=x;
  result->y=y; result->fontSize=size;
  result->color=*clr;
  return result;
}

/*
  Takes a BVG function call and calls the according C function
*/
void matchFunctionCall(char *call) {
  printf("Matching %s\n", call);
  char *funcCall = strdup(call);
  char *funcName = strsep(&funcCall, "(");
  char *function = trim(strlwr(funcName));
  free(funcName);
  printf("Got function %s\n", function);
  call[strlen(call)-1]='\0';
  if (strcmp(function, "rectangle") == 0) {
    char *argv[7];
    call = call+strlen("rectangle (");
    collectArgs(argv, call, 7);
    BVGRectangle *rect = malloc(sizeof(BVGRectangle)); 
    rect = BVGParseRectangle(argv);
    BVGDrawRectangle(rect);
  } else if (strcmp(function, "roundedrectangle") == 0) {
    char *argv[9];
    call = call+strlen("roundedrectangle (");
    collectArgs(argv, call, 9);
    BVGRoundedRectangle *roundrect = malloc(sizeof(BVGRoundedRectangle));
    roundrect = BVGParseRoundedRectangle(argv);
    BVGDrawRoundedRectangle(roundrect);
  } else if (strcmp(function, "circle") == 0) {
    char *argv[4];
    call = call+strlen("circle (");
    collectArgs(argv, call, 4);
    BVGCircle *circle = malloc(sizeof(BVGCircle));
    circle = BVGParseCircle(argv);
    BVGDrawCircle(circle);
  } else if (strcmp(function, "circlesegment") == 0) {
    char *argv[7];
    call = call+strlen("circlesegment (");
    collectArgs(argv, call, 7);
    BVGCircle *circle = malloc(sizeof(BVGCircle));
    circle = BVGParseCircleSegment(argv);
    BVGDrawCircle(circle);
  } else if (strcmp(function, "ring") == 0) {
    char *argv[8];
    call = call+strlen("ring (");
    collectArgs(argv, call, 8);
    BVGRing *ring = malloc(sizeof(BVGRing));
    ring = BVGParseRing(argv);
    BVGDrawRing(ring);
  } else if (strcmp(function, "text") == 0) {
    char *argv[5];
    call = call+strlen("text (");
    collectArgs(argv, call, 5);
    BVGText *text = malloc(sizeof(BVGText));
    text = BVGParseText(argv);
    BVGDrawText(text);
  }
  return;
}
