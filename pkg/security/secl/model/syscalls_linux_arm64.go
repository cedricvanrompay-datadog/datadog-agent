// Code generated - DO NOT EDIT.
// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux
// +build linux

package model

// Syscall represents a syscall identifier
type Syscall int

// Linux syscall identifiers
const (
	SysRestartSyscall           Syscall = 0
	SysExit                     Syscall = 1
	SysFork                     Syscall = 2
	SysRead                     Syscall = 3
	SysWrite                    Syscall = 4
	SysOpen                     Syscall = 5
	SysClose                    Syscall = 6
	SysCreat                    Syscall = 8
	SysLink                     Syscall = 9
	SysUnlink                   Syscall = 10
	SysExecve                   Syscall = 11
	SysChdir                    Syscall = 12
	SysTime                     Syscall = 13
	SysMknod                    Syscall = 14
	SysChmod                    Syscall = 15
	SysLchown                   Syscall = 16
	SysLseek                    Syscall = 19
	SysGetpid                   Syscall = 20
	SysMount                    Syscall = 21
	SysUmount                   Syscall = 22
	SysSetuid                   Syscall = 23
	SysGetuid                   Syscall = 24
	SysStime                    Syscall = 25
	SysPtrace                   Syscall = 26
	SysAlarm                    Syscall = 27
	SysPause                    Syscall = 29
	SysUtime                    Syscall = 30
	SysAccess                   Syscall = 33
	SysNice                     Syscall = 34
	SysSync                     Syscall = 36
	SysKill                     Syscall = 37
	SysRename                   Syscall = 38
	SysMkdir                    Syscall = 39
	SysRmdir                    Syscall = 40
	SysDup                      Syscall = 41
	SysPipe                     Syscall = 42
	SysTimes                    Syscall = 43
	SysBrk                      Syscall = 45
	SysSetgid                   Syscall = 46
	SysGetgid                   Syscall = 47
	SysGeteuid                  Syscall = 49
	SysGetegid                  Syscall = 50
	SysAcct                     Syscall = 51
	SysUmount2                  Syscall = 52
	SysIoctl                    Syscall = 54
	SysFcntl                    Syscall = 55
	SysSetpgid                  Syscall = 57
	SysUmask                    Syscall = 60
	SysChroot                   Syscall = 61
	SysUstat                    Syscall = 62
	SysDup2                     Syscall = 63
	SysGetppid                  Syscall = 64
	SysGetpgrp                  Syscall = 65
	SysSetsid                   Syscall = 66
	SysSigaction                Syscall = 67
	SysSetreuid                 Syscall = 70
	SysSetregid                 Syscall = 71
	SysSigsuspend               Syscall = 72
	SysSigpending               Syscall = 73
	SysSethostname              Syscall = 74
	SysSetrlimit                Syscall = 75
	SysGetrlimit                Syscall = 76
	SysGetrusage                Syscall = 77
	SysGettimeofday             Syscall = 78
	SysSettimeofday             Syscall = 79
	SysGetgroups                Syscall = 80
	SysSetgroups                Syscall = 81
	SysSelect                   Syscall = 82
	SysSymlink                  Syscall = 83
	SysReadlink                 Syscall = 85
	SysUselib                   Syscall = 86
	SysSwapon                   Syscall = 87
	SysReboot                   Syscall = 88
	SysReaddir                  Syscall = 89
	SysMmap                     Syscall = 90
	SysMunmap                   Syscall = 91
	SysTruncate                 Syscall = 92
	SysFtruncate                Syscall = 93
	SysFchmod                   Syscall = 94
	SysFchown                   Syscall = 95
	SysGetpriority              Syscall = 96
	SysSetpriority              Syscall = 97
	SysStatfs                   Syscall = 99
	SysFstatfs                  Syscall = 100
	SysSocketcall               Syscall = 102
	SysSyslog                   Syscall = 103
	SysSetitimer                Syscall = 104
	SysGetitimer                Syscall = 105
	SysStat                     Syscall = 106
	SysLstat                    Syscall = 107
	SysFstat                    Syscall = 108
	SysVhangup                  Syscall = 111
	SysSyscall                  Syscall = 113
	SysWait4                    Syscall = 114
	SysSwapoff                  Syscall = 115
	SysSysinfo                  Syscall = 116
	SysIpc                      Syscall = 117
	SysFsync                    Syscall = 118
	SysSigreturn                Syscall = 119
	SysClone                    Syscall = 120
	SysSetdomainname            Syscall = 121
	SysUname                    Syscall = 122
	SysAdjtimex                 Syscall = 124
	SysMprotect                 Syscall = 125
	SysSigprocmask              Syscall = 126
	SysInitModule               Syscall = 128
	SysDeleteModule             Syscall = 129
	SysQuotactl                 Syscall = 131
	SysGetpgid                  Syscall = 132
	SysFchdir                   Syscall = 133
	SysBdflush                  Syscall = 134
	SysSysfs                    Syscall = 135
	SysPersonality              Syscall = 136
	SysSetfsuid                 Syscall = 138
	SysSetfsgid                 Syscall = 139
	SysLlseek                   Syscall = 140
	SysGetdents                 Syscall = 141
	SysNewselect                Syscall = 142
	SysFlock                    Syscall = 143
	SysMsync                    Syscall = 144
	SysReadv                    Syscall = 145
	SysWritev                   Syscall = 146
	SysGetsid                   Syscall = 147
	SysFdatasync                Syscall = 148
	SysSysctl                   Syscall = 149
	SysMlock                    Syscall = 150
	SysMunlock                  Syscall = 151
	SysMlockall                 Syscall = 152
	SysMunlockall               Syscall = 153
	SysSchedSetparam            Syscall = 154
	SysSchedGetparam            Syscall = 155
	SysSchedSetscheduler        Syscall = 156
	SysSchedGetscheduler        Syscall = 157
	SysSchedYield               Syscall = 158
	SysSchedGetPriorityMax      Syscall = 159
	SysSchedGetPriorityMin      Syscall = 160
	SysSchedRrGetInterval       Syscall = 161
	SysNanosleep                Syscall = 162
	SysMremap                   Syscall = 163
	SysSetresuid                Syscall = 164
	SysGetresuid                Syscall = 165
	SysPoll                     Syscall = 168
	SysNfsservctl               Syscall = 169
	SysSetresgid                Syscall = 170
	SysGetresgid                Syscall = 171
	SysPrctl                    Syscall = 172
	SysRtSigreturn              Syscall = 173
	SysRtSigaction              Syscall = 174
	SysRtSigprocmask            Syscall = 175
	SysRtSigpending             Syscall = 176
	SysRtSigtimedwait           Syscall = 177
	SysRtSigqueueinfo           Syscall = 178
	SysRtSigsuspend             Syscall = 179
	SysPread64                  Syscall = 180
	SysPwrite64                 Syscall = 181
	SysChown                    Syscall = 182
	SysGetcwd                   Syscall = 183
	SysCapget                   Syscall = 184
	SysCapset                   Syscall = 185
	SysSigaltstack              Syscall = 186
	SysSendfile                 Syscall = 187
	SysVfork                    Syscall = 190
	SysUgetrlimit               Syscall = 191
	SysMmap2                    Syscall = 192
	SysTruncate64               Syscall = 193
	SysFtruncate64              Syscall = 194
	SysStat64                   Syscall = 195
	SysLstat64                  Syscall = 196
	SysFstat64                  Syscall = 197
	SysLchown32                 Syscall = 198
	SysGetuid32                 Syscall = 199
	SysGetgid32                 Syscall = 200
	SysGeteuid32                Syscall = 201
	SysGetegid32                Syscall = 202
	SysSetreuid32               Syscall = 203
	SysSetregid32               Syscall = 204
	SysGetgroups32              Syscall = 205
	SysSetgroups32              Syscall = 206
	SysFchown32                 Syscall = 207
	SysSetresuid32              Syscall = 208
	SysGetresuid32              Syscall = 209
	SysSetresgid32              Syscall = 210
	SysGetresgid32              Syscall = 211
	SysChown32                  Syscall = 212
	SysSetuid32                 Syscall = 213
	SysSetgid32                 Syscall = 214
	SysSetfsuid32               Syscall = 215
	SysSetfsgid32               Syscall = 216
	SysGetdents64               Syscall = 217
	SysPivotRoot                Syscall = 218
	SysMincore                  Syscall = 219
	SysMadvise                  Syscall = 220
	SysFcntl64                  Syscall = 221
	SysGettid                   Syscall = 224
	SysReadahead                Syscall = 225
	SysSetxattr                 Syscall = 226
	SysLsetxattr                Syscall = 227
	SysFsetxattr                Syscall = 228
	SysGetxattr                 Syscall = 229
	SysLgetxattr                Syscall = 230
	SysFgetxattr                Syscall = 231
	SysListxattr                Syscall = 232
	SysLlistxattr               Syscall = 233
	SysFlistxattr               Syscall = 234
	SysRemovexattr              Syscall = 235
	SysLremovexattr             Syscall = 236
	SysFremovexattr             Syscall = 237
	SysTkill                    Syscall = 238
	SysSendfile64               Syscall = 239
	SysFutex                    Syscall = 240
	SysSchedSetaffinity         Syscall = 241
	SysSchedGetaffinity         Syscall = 242
	SysIoSetup                  Syscall = 243
	SysIoDestroy                Syscall = 244
	SysIoGetevents              Syscall = 245
	SysIoSubmit                 Syscall = 246
	SysIoCancel                 Syscall = 247
	SysExitGroup                Syscall = 248
	SysLookupDcookie            Syscall = 249
	SysEpollCreate              Syscall = 250
	SysEpollCtl                 Syscall = 251
	SysEpollWait                Syscall = 252
	SysRemapFilePages           Syscall = 253
	SysSetTidAddress            Syscall = 256
	SysTimerCreate              Syscall = 257
	SysTimerSettime             Syscall = 258
	SysTimerGettime             Syscall = 259
	SysTimerGetoverrun          Syscall = 260
	SysTimerDelete              Syscall = 261
	SysClockSettime             Syscall = 262
	SysClockGettime             Syscall = 263
	SysClockGetres              Syscall = 264
	SysClockNanosleep           Syscall = 265
	SysStatfs64                 Syscall = 266
	SysFstatfs64                Syscall = 267
	SysTgkill                   Syscall = 268
	SysUtimes                   Syscall = 269
	SysArmFadvise6464           Syscall = 270
	SysPciconfigIobase          Syscall = 271
	SysPciconfigRead            Syscall = 272
	SysPciconfigWrite           Syscall = 273
	SysMqOpen                   Syscall = 274
	SysMqUnlink                 Syscall = 275
	SysMqTimedsend              Syscall = 276
	SysMqTimedreceive           Syscall = 277
	SysMqNotify                 Syscall = 278
	SysMqGetsetattr             Syscall = 279
	SysWaitid                   Syscall = 280
	SysSocket                   Syscall = 281
	SysBind                     Syscall = 282
	SysConnect                  Syscall = 283
	SysListen                   Syscall = 284
	SysAccept                   Syscall = 285
	SysGetsockname              Syscall = 286
	SysGetpeername              Syscall = 287
	SysSocketpair               Syscall = 288
	SysSend                     Syscall = 289
	SysSendto                   Syscall = 290
	SysRecv                     Syscall = 291
	SysRecvfrom                 Syscall = 292
	SysShutdown                 Syscall = 293
	SysSetsockopt               Syscall = 294
	SysGetsockopt               Syscall = 295
	SysSendmsg                  Syscall = 296
	SysRecvmsg                  Syscall = 297
	SysSemop                    Syscall = 298
	SysSemget                   Syscall = 299
	SysSemctl                   Syscall = 300
	SysMsgsnd                   Syscall = 301
	SysMsgrcv                   Syscall = 302
	SysMsgget                   Syscall = 303
	SysMsgctl                   Syscall = 304
	SysShmat                    Syscall = 305
	SysShmdt                    Syscall = 306
	SysShmget                   Syscall = 307
	SysShmctl                   Syscall = 308
	SysAddKey                   Syscall = 309
	SysRequestKey               Syscall = 310
	SysKeyctl                   Syscall = 311
	SysSemtimedop               Syscall = 312
	SysVserver                  Syscall = 313
	SysIoprioSet                Syscall = 314
	SysIoprioGet                Syscall = 315
	SysInotifyInit              Syscall = 316
	SysInotifyAddWatch          Syscall = 317
	SysInotifyRmWatch           Syscall = 318
	SysMbind                    Syscall = 319
	SysGetMempolicy             Syscall = 320
	SysSetMempolicy             Syscall = 321
	SysOpenat                   Syscall = 322
	SysMkdirat                  Syscall = 323
	SysMknodat                  Syscall = 324
	SysFchownat                 Syscall = 325
	SysFutimesat                Syscall = 326
	SysFstatat64                Syscall = 327
	SysUnlinkat                 Syscall = 328
	SysRenameat                 Syscall = 329
	SysLinkat                   Syscall = 330
	SysSymlinkat                Syscall = 331
	SysReadlinkat               Syscall = 332
	SysFchmodat                 Syscall = 333
	SysFaccessat                Syscall = 334
	SysPselect6                 Syscall = 335
	SysPpoll                    Syscall = 336
	SysUnshare                  Syscall = 337
	SysSetRobustList            Syscall = 338
	SysGetRobustList            Syscall = 339
	SysSplice                   Syscall = 340
	SysArmSyncFileRange         Syscall = 341
	SysTee                      Syscall = 342
	SysVmsplice                 Syscall = 343
	SysMovePages                Syscall = 344
	SysGetcpu                   Syscall = 345
	SysEpollPwait               Syscall = 346
	SysKexecLoad                Syscall = 347
	SysUtimensat                Syscall = 348
	SysSignalfd                 Syscall = 349
	SysTimerfdCreate            Syscall = 350
	SysEventfd                  Syscall = 351
	SysFallocate                Syscall = 352
	SysTimerfdSettime           Syscall = 353
	SysTimerfdGettime           Syscall = 354
	SysSignalfd4                Syscall = 355
	SysEventfd2                 Syscall = 356
	SysEpollCreate1             Syscall = 357
	SysDup3                     Syscall = 358
	SysPipe2                    Syscall = 359
	SysInotifyInit1             Syscall = 360
	SysPreadv                   Syscall = 361
	SysPwritev                  Syscall = 362
	SysRtTgsigqueueinfo         Syscall = 363
	SysPerfEventOpen            Syscall = 364
	SysRecvmmsg                 Syscall = 365
	SysAccept4                  Syscall = 366
	SysFanotifyInit             Syscall = 367
	SysFanotifyMark             Syscall = 368
	SysPrlimit64                Syscall = 369
	SysNameToHandleAt           Syscall = 370
	SysOpenByHandleAt           Syscall = 371
	SysClockAdjtime             Syscall = 372
	SysSyncfs                   Syscall = 373
	SysSendmmsg                 Syscall = 374
	SysSetns                    Syscall = 375
	SysProcessVmReadv           Syscall = 376
	SysProcessVmWritev          Syscall = 377
	SysKcmp                     Syscall = 378
	SysFinitModule              Syscall = 379
	SysSchedSetattr             Syscall = 380
	SysSchedGetattr             Syscall = 381
	SysRenameat2                Syscall = 382
	SysSeccomp                  Syscall = 383
	SysGetrandom                Syscall = 384
	SysMemfdCreate              Syscall = 385
	SysBpf                      Syscall = 386
	SysExecveat                 Syscall = 387
	SysUserfaultfd              Syscall = 388
	SysMembarrier               Syscall = 389
	SysMlock2                   Syscall = 390
	SysCopyFileRange            Syscall = 391
	SysPreadv2                  Syscall = 392
	SysPwritev2                 Syscall = 393
	SysPkeyMprotect             Syscall = 394
	SysPkeyAlloc                Syscall = 395
	SysPkeyFree                 Syscall = 396
	SysStatx                    Syscall = 397
	SysRseq                     Syscall = 398
	SysIoPgetevents             Syscall = 399
	SysMigratePages             Syscall = 400
	SysKexecFileLoad            Syscall = 401
	SysClockGettime64           Syscall = 403
	SysClockSettime64           Syscall = 404
	SysClockAdjtime64           Syscall = 405
	SysClockGetresTime64        Syscall = 406
	SysClockNanosleepTime64     Syscall = 407
	SysTimerGettime64           Syscall = 408
	SysTimerSettime64           Syscall = 409
	SysTimerfdGettime64         Syscall = 410
	SysTimerfdSettime64         Syscall = 411
	SysUtimensatTime64          Syscall = 412
	SysPselect6Time64           Syscall = 413
	SysPpollTime64              Syscall = 414
	SysIoPgeteventsTime64       Syscall = 416
	SysRecvmmsgTime64           Syscall = 417
	SysMqTimedsendTime64        Syscall = 418
	SysMqTimedreceiveTime64     Syscall = 419
	SysSemtimedopTime64         Syscall = 420
	SysRtSigtimedwaitTime64     Syscall = 421
	SysFutexTime64              Syscall = 422
	SysSchedRrGetIntervalTime64 Syscall = 423
	SysPidfdSendSignal          Syscall = 424
	SysIoUringSetup             Syscall = 425
	SysIoUringEnter             Syscall = 426
	SysIoUringRegister          Syscall = 427
	SysOpenTree                 Syscall = 428
	SysMoveMount                Syscall = 429
	SysFsopen                   Syscall = 430
	SysFsconfig                 Syscall = 431
	SysFsmount                  Syscall = 432
	SysFspick                   Syscall = 433
	SysPidfdOpen                Syscall = 434
	SysClone3                   Syscall = 435
	SysCloseRange               Syscall = 436
	SysOpenat2                  Syscall = 437
	SysPidfdGetfd               Syscall = 438
	SysFaccessat2               Syscall = 439
	SysProcessMadvise           Syscall = 440
	SysEpollPwait2              Syscall = 441
	SysMountSetattr             Syscall = 442
	SysQuotactlFd               Syscall = 443
	SysLandlockCreateRuleset    Syscall = 444
	SysLandlockAddRule          Syscall = 445
	SysLandlockRestrictSelf     Syscall = 446
	SysProcessMrelease          Syscall = 448
	SysFutexWaitv               Syscall = 449
	SysSetMempolicyHomeNode     Syscall = 450
)
