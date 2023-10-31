# ./tests/hello-world-server.nix
(import ./lib.nix) {
  name = "gokill-base-test";
  nodes = {
    # `self` here is set by using specialArgs in `lib.nix`
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
                  type = "Command";
                  options = {
                      command = "echo hello world";
                  };
                  stage = 2;
              }
            ];
          }
        ];
      };
    };
  };

  testScript = ''
    import time
    start_all() # wait for our service to start
    node1.wait_for_unit("gokill")
    time.sleep(11)
    output = node1.succeed("journalctl -u gokill.service | tail -n 2 | head -n 1")
    # Check if our webserver returns the expected result
    assert "hellow world" in output
  '';
}
