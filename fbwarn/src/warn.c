#include "raylib.h"
#include "parseBVG.h"
#include <string.h>
#include <stdlib.h>
#include <unistd.h>

int main(void) {
  char *rectA, *rectB, *rectC, *rectAFree, *rectBFree, *rectCFree, *singleA, *singleB, *singleC;
  rectAFree = rectA = strdup("rectangle (x=0,y=0,\nwidth=100,\nheight=100,\ncolor='#5BCEFA')\n");
  if (rectA == NULL)
    return 1;
  singleA = multiToSingle(rectA);
  free(rectAFree);

  rectBFree = rectB = strdup("rectangle (x=0,y=20,\nwidth=100,\nheight=60,\ncolor='#F5A9B8')\n");
  if (rectB == NULL)
    return 1;
  singleB = multiToSingle(rectB);
  free(rectBFree);

  rectCFree = rectC = strdup("rectangle (x=0,y=40,\nwidth=100,\nheight=20,\ncolor='#FFFFFF')\n");
  if (rectC == NULL)
    return 1;
  singleC = multiToSingle(rectC);
  free(rectCFree);
  
  InitWindow (100, 100, ":3");

  while (!WindowShouldClose ()) {
    BeginDrawing ();
    ClearBackground (RAYWHITE);

    char *parseA = strdup(singleA);
    matchFunctionCall(parseA);
    
    char *parseB = strdup(singleB);
    matchFunctionCall(parseB);
    
    char *parseC = strdup(singleC);
    matchFunctionCall(parseC);
    DrawText ("tranmsgenmer", 15, 50, 10, MAROON);
    EndDrawing ();
  }

  CloseWindow ();
  
  return 0;
}
