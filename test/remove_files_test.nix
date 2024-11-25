{ pkgs, ... }:
(import ./lib.nix) {
  name = "gokill-remove-files-test";
  nodes = {
    node1 = { self, pkgs, ... }: {
      imports = [ self.nixosModules.gokill ];

      services.gokill = {
        enable = true;
        triggers = [
          {
            type = "Timeout";
            name = "custom timeout";
            options = {
              duration =  10;
            };
            actions = [
              {
                  type = "RemoveFiles";
                  options = {
                    files = [
                      "/tmp/file1.txt"
                      "/tmp/file2.txt"
                      "/tmp/file3.txt"
                      "/tmp/file4.txt"
                      "/tmp/file5.txt" # does not exist
                    ];
                    directories = [
                      "/tmp/dir1"
                      "/tmp/dir2"
                      "/tmp/dir3"
                      "/tmp/dir4"
                      "/tmp/dir5" # does not exist
                    ];
                  };
                  stage = 2;
              }
            ];
          }
        ];
      };
    };
  };

  testScript = let
    createTestData = pkgs.writeScript "createTestData" ''
      echo "1" > /tmp/file1.txt
      echo "2" > /tmp/file2.txt
      echo "3" > /tmp/file3.txt
      echo "4" > /tmp/file4.txt
      mkdir /tmp/dir1
      mkdir /tmp/dir2
      mkdir /tmp/dir3
      mkdir /tmp/dir4
    '';
  in ''
    import time

    def run(command):
      status, stdout = node1.execute(command)
      print(stdout)

    start_all() # wait for our service to start
    node1.wait_for_unit("gokill")

    run("${createTestData}")

    run("ls -la /tmp")
    run("ls -la /")


    #node1.succeed("systemctl status gokill.service")
    run("journalctl -u gokill.service")

    filenames = [
      "/tmp/file1.txt",
      "/tmp/file2.txt",
      "/tmp/file3.txt",
      "/tmp/file4.txt",
      "/tmp/dir1",
      "/tmp/dir2",
      "/tmp/dir3",
      "/tmp/dir4"
    ];

    for name in filenames:
      node1.succeed("test -e " + name + " && (exit 0) || (exit 1)")

    time.sleep(60)
    run("ls -la /tmp")

    for name in filenames:
      node1.fail("test -e " + name + " && (exit 0) || (exit 1)")
  '';
}
