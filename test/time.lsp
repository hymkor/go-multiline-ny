; https://github.com/hymkor/lispect

(defglobal pid (spawn "go" "run" "slow.go"))
(expect "[1]")
(defglobal start (get-internal-real-time))
(sendln "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890")
(send #\UA)
(send #\U4)
(wait pid)
(defglobal end (get-internal-real-time))
(format (standard-output) "TIME: ~a~%" (div (convert (- end start) <float>) (internal-time-units-per-second)))
