# krononlabs
언어: 

- python 또는 golang 중 한 언어를 선택해 구현해주세요.

Database: 

- 자유 (단, docker-compose로 환경을 구성해주세요.)

1. 각 거래소 API, Websocket을 사용하여 데이터를 가져오는 코드를 작성해주세요.
    1. 오픈소스를 사용하지 말고, 각 거래소의 API 문서를 참고하여 직접 개발해주세요. 
    2. 사용할 거래소 (선택이 아닌 둘다 사용하여 개발해주세요.)
        1. binance
        2. bitget
    3. 거래쌍 
        1. BTC-USDT Perpetual
        2. ETH-USDT Perpetual
        3. XRP-USDT Perpetual
    - 구현 내용
        - KLINE 조회 API
        - KLINE 및 TRADE Websocket
2. **[첨부]** 에 제공해드린 전략함수를 동작시키는 로직을 추가해주세요. 
    - 전략함수의 동작은 BACK / FORWARD 2가지 버전을 작성해주세요.
        - BACK (”과거 데이터를 가져올 수 있는 레퍼런스”를 참고해주세요.)
            - 과거의 데이터로 전략함수를 동작
        - FORWARD (”1번 과제에서 작성한 코드를 사용해주세요.”)
            - 실시간 데이터를 기반으로 전략함수를 동작
3.  전략함수 결과값을 Database에 저장하는 로직을 작성해주세요.
    1. `test_type` Column을 추가하여 BACK / FORWARD 를 구분해주세요.
        1. 이 외 Column은 자유롭게 작성해주시면 됩니다. 
    2. BACK 데이터는 과제일자 기준 이전 데이터를 편하신 기간을 정하여 사용해주세요.
    3. FORWARD 데이터는 API 및 Websocket을 사용하여 실시간 데이터를 사용해주세요.
