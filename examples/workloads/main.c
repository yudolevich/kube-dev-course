#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>

int main(int argc, char *argv[]) {
  if (argc < 3) {
    printf("too few arguments\n");
    return 1;
  }

  while (1) {
    int fd = open(argv[1], O_RDONLY);
    if (fd < 0) {
      printf("error read file: %s\n", argv[1]);
      sleep(strtol(argv[2], NULL, 10));
      continue;
    }
    struct stat s;
    int status = fstat(fd, &s);
    char *data = mmap(0, s.st_size, PROT_READ, MAP_SHARED, fd, 0);
    for (int i = 0; i < s.st_size; i++) {
      char c;
      c = data[i];
    }
    printf("size: %lld\n", s.st_size);
    sleep(strtol(argv[2], NULL, 10));
    munmap(data, s.st_size);
    close(fd);
  }
}
