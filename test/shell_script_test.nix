{ pkgs, ... }:
(import ./lib.nix) {
  name = "gokill-remove-files-test";
  nodes = {
    node1 = { self, pkgs, ... }: let
      simpleTestScript = pkgs.writeScript "simpleTestScript" ''
        echo "hello world"
      '';
    in {
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
                  type = "ShellScript";
                  options = {
                    path = "${simpleTestScript}";
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
  in ''
    import time

    start_all() # wait for our service to start
    node1.wait_for_unit("gokill")
    time.sleep(5)
    output = node1.succeed("journalctl -u gokill.service | tail -n 20")
    assert "hello world" in output
  '';
}
