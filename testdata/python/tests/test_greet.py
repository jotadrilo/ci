from greet import hello, newer


def test_hello():
    assert hello("CI") == "Hello, CI!"


def test_newer():
    assert newer("1.10", "1.9") == "1.10"
