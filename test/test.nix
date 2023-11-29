(import ./lib.nix) {
  name = "gokill-base-test";
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
              duration =  3;
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
    time.sleep(4)
    output = node1.succeed("journalctl -u gokill.service | tail -n 20")
    # Check if our webserver returns the expected result
    assert "hello world" in output
  '';
}
