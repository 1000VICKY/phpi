<?php 
phpinfo();
echo(PHP_EOL);print_r($_SERVER);
echo(PHP_EOL);$res = File_geT_contents("http://nasse.com");
echo(PHP_EOL);echo ($res);
echo(PHP_EOL);