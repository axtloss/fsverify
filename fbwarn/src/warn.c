#include "raylib.h"
#include "parseBVG.h"
#include <string.h>
#include <stdlib.h>
#include <unistd.h>

int main(void) {
  char *rectA, *rectB, *rectC, *rectD, *rectAFree, *rectBFree, *rectCFree, *rectDFree, *singleA, *singleB, *singleC, *singleD;
  char *rectE, *rectF, *rectEFree, *rectFFree, *singleE, *singleF;
  rectAFree = rectA = strdup("rectangle (x=10,y=20,\nwidth=100,\nheight=100,\ncolor='#9787CFFF',\nfill=true,\nthickness=1.0)\n");
  if (rectA == NULL)
    return 1;
  singleA = multiToSingle(rectA);
  free(rectAFree);

  rectBFree = rectB = strdup("rectangle (x=130,y=160,\nwidth=100,\nheight=60,\ncolor='#88C2B1FF',\nfill=false,\nthickness=5.0)\n");
  if (rectB == NULL)
    return 1;
  singleB = multiToSingle(rectB);
  free(rectBFree);

  rectCFree = rectC = strdup("circlesegment (x=300, y=200,radius=100,color='#BE79A7FF',startangle=0.0,endangle=90.0,segments=10)");
  if (rectC == NULL)
    return 1;
  singleC = multiToSingle(rectC);
  printf("SingleC %s", rectC);
  free(rectCFree);

  rectDFree = rectD = strdup("ring (x=300,y=50,innerradius=20,outerradius=30,startangle=0.0,endangle=360.0,segments=10,color='#DD98E5FF')");
  //  rectDFree = rectD = strdup("text (text='haiii :3',x=300,y=10,size=50,color='#DD98E5FF')");
  if (rectD == NULL)
    return 1;
  singleD = multiToSingle(rectD);
  free(rectDFree);

  rectEFree = rectE = strdup("roundedrectangle (x=90,y=300,\nwidth=92,\nheight=20,\ncolor='#BE79A7FF',\nfill=false,\nthickness=3.0,roundness=5.0,segments=100)\n");
  if (rectE == NULL)
    return 1;
  singleE = multiToSingle(rectE);
  free(rectEFree);

  rectFFree = rectF = strdup("circle (x=700,y=300,radius=90.0,color='#7676DCFF')\n");
  if (rectF == NULL)
    return 1;
  singleF = multiToSingle(rectF);
  free(rectFFree);
  
  
  InitWindow (800, 400, ":3");


  while (!WindowShouldClose ()) {
    
    BeginDrawing ();
    ClearBackground (RAYWHITE);
    char *parseA = strdup(singleA);
    matchFunctionCall(parseA);
  
    char *parseB = strdup(singleB);
    matchFunctionCall(parseB);
  
    char *parseC = strdup(singleC);
    matchFunctionCall(parseC);

    char *parseD = strdup(singleD);
    matchFunctionCall(parseD);
    
    char *parseE = strdup(singleE);
    matchFunctionCall(parseE);

    char *parseF = strdup(singleF);
    matchFunctionCall(parseF);
    
    char *text = malloc(strlen("100")*100);
    sprintf(text, "%d", GetFPS());
    DrawText(text, 2, 2, 20, MAROON);
    EndDrawing ();
  }

  CloseWindow ();
  
  return 0;
}
