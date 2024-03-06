#include <stdlib.h>
#include <unistd.h>
#include <stdio.h>
#include <linux/fb.h>
#include <sys/ioctl.h>
#include <fcntl.h>


int main(void) {
  int fbfb = open("/dev/fb0", O_RDONLY);
  struct fb_var_screeninfo vinfo;  

  printf("%d", 1920);
  return 0;
  
  if (fbfb < 0)
    return 1;

  if (ioctl(fbfb, FBIOGET_VSCREENINFO, &vinfo) == -1)
    return 1;

  printf("%d", vinfo.xres);
  return 0;
}
