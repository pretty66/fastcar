<?php

function request($url, $method = 'GET', string $data = '', $headers = [])
{
    $method = strtoupper($method);
    $ch = curl_init();
    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);
    $method === 'POST' && curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
    empty($headers) || curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

    /******* add option *******/
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
    curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);
    curl_setopt($ch, CURLOPT_UNIX_SOCKET_PATH, '/tmp/fastcar.sock');
    /*************************/

    $t1 = microtime(true);
    $response = curl_exec($ch);
    $delay = microtime(true) - $t1;
    // 检查是否有错误发生
    if (curl_errno($ch)) {
        return [
            'error' => curl_error($ch)
        ];
    }

    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

    curl_close($ch);
    return [
        'status'   => $httpCode,
        'response' => $response,
        'delay'    => round($delay * 1000, 2),
    ];
}

for ($i = 0; $i < 100; $i++) {
    sleep(1);
    // $res = request('http://api.fanyi.baidu.com/api/trans/vip/translate');
    $res = request('https://api.fanyi.baidu.com/api/trans/vip/translate');
    $delay = $res['delay'];
    echo $delay . PHP_EOL;
}


