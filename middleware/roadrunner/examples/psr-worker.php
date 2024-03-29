<?php

use Spiral\RoadRunner;
use Nyholm\Psr7;

include "vendor/autoload.php";

$worker = RoadRunner\Worker::create();
$psrFactory = new Psr7\Factory\Psr17Factory();

$worker = new RoadRunner\Http\PSR7Worker($worker, $psrFactory, $psrFactory, $psrFactory);

if (file_exists('break')) {
	throw new Exception('oops');
}

while ($req = $worker->waitRequest()) {
    try {
        $rsp = new Psr7\Response();
        if ($_SERVER['REQUEST_URI'] === 'http://localhost/include') {
            $rsp->getBody()->write('Include content');
        } else {
            $rsp->getBody()->write('Hello base uri! <esi:include src="/include" />!');
        }

        $worker->respond($rsp);
    } catch (\Throwable $e) {
        $worker->getWorker()->error((string)$e);
    }
}
