#include "raylib.h"
#include "parseBVG.h"
#include <string.h>
#include <stdlib.h>

int main(void) {
  char *str, *single, *toFree;
  toFree = str = strdup("rectangle (x=20,y=30,\nwidth=200,\nheight=300,\ncolor='#F5A9B8')\n");
  if (str == NULL)
    return 1;
  printf("Multi:\n %s", str);
  single = multiToSingle(str);
  printf("Single:\n %s\n", single);
  free(toFree);
    
  InitWindow (800, 450, "raylib");

  while (!WindowShouldClose ()) {
    BeginDrawing ();
    ClearBackground (RAYWHITE);
    DrawText ("TEXT", 190, 200, 20, LIGHTGRAY);
    char *parse = strdup(single);
    matchFunctionCall(parse);
    EndDrawing ();
  }

  CloseWindow ();
  
  return 0;
}
