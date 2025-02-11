#ifndef _FLOW_H_
#define _FLOW_H_

#define EGRESS 1
#define INGRESS 2

struct pid_route_t {
    u64 addr[2];
    u32 netns;
    u16 port;
};

struct bpf_map_def SEC("maps/flow_pid") flow_pid = {
    .type = BPF_MAP_TYPE_LRU_HASH,
    .key_size = sizeof(struct pid_route_t),
    .value_size = sizeof(u32),
    .max_entries = 10240,
    .pinning = 0,
    .namespace = "",
};

__attribute__((always_inline)) u32 get_flow_pid(struct pid_route_t *key) {
    u32 *value = bpf_map_lookup_elem(&flow_pid, key);
    if (!value) {
        // Try with IP set to 0.0.0.0
        key->addr[0] = 0;
        key->addr[1] = 0;
        value = bpf_map_lookup_elem(&flow_pid, key);
        if (!value) {
            return 0;
        }
    }

    return *value;
}

__attribute__((always_inline)) u16 get_family_from_sock_common(struct sock_common *sk) {
    u64 sock_common_skc_family_offset;
    LOAD_CONSTANT("sock_common_skc_family_offset", sock_common_skc_family_offset);

    u16 family;
    bpf_probe_read(&family, sizeof(family), (void*)sk + sock_common_skc_family_offset);
    return family;
}

__attribute__((always_inline)) u64 get_flowi4_saddr_offset() {
    u64 flowi4_saddr_offset;
    LOAD_CONSTANT("flowi4_saddr_offset", flowi4_saddr_offset);
    return flowi4_saddr_offset;
}

__attribute__((always_inline)) u64 get_flowi4_uli_offset() {
    u64 flowi4_uli_offset;
    LOAD_CONSTANT("flowi4_uli_offset", flowi4_uli_offset);
    return flowi4_uli_offset;
}

__attribute__((always_inline)) u64 get_flowi6_saddr_offset() {
    u64 flowi6_saddr_offset;
    LOAD_CONSTANT("flowi6_saddr_offset", flowi6_saddr_offset);
    return flowi6_saddr_offset;
}

__attribute__((always_inline)) u64 get_flowi6_uli_offset() {
    u64 flowi6_uli_offset;
    LOAD_CONSTANT("flowi6_uli_offset", flowi6_uli_offset);
    return flowi6_uli_offset;
}

SEC("kprobe/security_sk_classify_flow")
int kprobe_security_sk_classify_flow(struct pt_regs *ctx) {
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
    struct flowi *fl = (struct flowi *)PT_REGS_PARM2(ctx);
    struct pid_route_t key = {};
    union flowi_uli uli;

    u16 family = get_family_from_sock_common((void *)sk);
    if (family == AF_INET6) {
        bpf_probe_read(&key.addr, sizeof(u64) * 2, (void *)fl + get_flowi6_saddr_offset());
        bpf_probe_read(&uli, sizeof(uli), (void *)fl + get_flowi6_uli_offset());
        bpf_probe_read(&key.port, sizeof(key.port), &uli.ports.sport);
    } else if (family == AF_INET) {
        bpf_probe_read(&key.addr, sizeof(u32), (void *)fl + get_flowi4_saddr_offset());
        bpf_probe_read(&uli, sizeof(uli), (void *)fl + get_flowi4_uli_offset());
        bpf_probe_read(&key.port, sizeof(key.port), &uli.ports.sport);
    } else {
        return 0;
    }

    // Register service PID
    if (key.port != 0) {
        u64 id = bpf_get_current_pid_tgid();
        u32 tid = (u32)id;
        u32 pid = id >> 32;

        // add netns information
        key.netns = get_netns_from_sock(sk);
        if (key.netns != 0) {
            bpf_map_update_elem(&netns_cache, &tid, &key.netns, BPF_ANY);
        }

        bpf_map_update_elem(&flow_pid, &key, &pid, BPF_ANY);

#ifdef DEBUG
        bpf_printk("# registered (flow) pid:%d netns:%u\n", pid, key.netns);
        bpf_printk("# p:%d a:%d a:%d\n", key.port, key.addr[0], key.addr[1]);
#endif
    }
    return 0;
}

SEC("kprobe/security_socket_bind")
int kprobe_security_socket_bind(struct pt_regs *ctx) {
    struct socket *sk = (struct socket *)PT_REGS_PARM1(ctx);
    struct sockaddr *address = (struct sockaddr *)PT_REGS_PARM2(ctx);
    struct pid_route_t key = {};
    u16 family = 0;

    // Extract IP and port from the sockaddr structure
    bpf_probe_read(&family, sizeof(family), &address->sa_family);
    if (family == AF_INET) {
        struct sockaddr_in *addr_in = (struct sockaddr_in *)address;
        bpf_probe_read(&key.port, sizeof(addr_in->sin_port), &addr_in->sin_port);
        bpf_probe_read(&key.addr, sizeof(addr_in->sin_addr.s_addr), &addr_in->sin_addr.s_addr);
    } else if (family == AF_INET6) {
        struct sockaddr_in6 *addr_in6 = (struct sockaddr_in6 *)address;
        bpf_probe_read(&key.port, sizeof(addr_in6->sin6_port), &addr_in6->sin6_port);
        bpf_probe_read(&key.addr, sizeof(u64) * 2, (char *)addr_in6 + offsetof(struct sockaddr_in6, sin6_addr));
    }

    // fill syscall_cache if necessary
    struct syscall_cache_t *syscall = peek_syscall(EVENT_BIND);
    if (syscall) {
        syscall->bind.addr[0] = key.addr[0];
        syscall->bind.addr[1] = key.addr[1];
        syscall->bind.port = key.port;
        syscall->bind.family = family;
    }

    // past this point we care only about AF_INET and AF_INET6
    if (family != AF_INET && family != AF_INET6) {
        return 0;
    }

    // Register service PID
    if (key.port != 0) {
        u64 id = bpf_get_current_pid_tgid();
        u32 tid = (u32) id;
        u32 pid = id >> 32;

        // add netns information
        key.netns = get_netns_from_socket(sk);
        if (key.netns != 0) {
            bpf_map_update_elem(&netns_cache, &tid, &key.netns, BPF_ANY);
        }

        bpf_map_update_elem(&flow_pid, &key, &pid, BPF_ANY);

#ifdef DEBUG
        bpf_printk("# registered (bind) pid:%d\n", pid);
        bpf_printk("# p:%d a:%d a:%d\n", key.port, key.addr[0], key.addr[1]);
#endif
    }
    return 0;
}

struct flow_t {
    u64 saddr[2];
    u64 daddr[2];
    u16 sport;
    u16 dport;
    u32 padding;
};

struct namespaced_flow_t {
    struct flow_t flow;
    u32 netns;
};

__attribute__((always_inline)) void flip(struct flow_t *flow) {
    u64 tmp = 0;
    tmp = flow->sport;
    flow->sport = flow->dport;
    flow->dport = tmp;

    tmp = flow->saddr[0];
    flow->saddr[0] = flow->daddr[0];
    flow->daddr[0] = tmp;

    tmp = flow->saddr[1];
    flow->saddr[1] = flow->daddr[1];
    flow->daddr[1] = tmp;
}

__attribute__((always_inline)) void parse_tuple(struct nf_conntrack_tuple *tuple, struct flow_t *flow) {
    flow->sport = tuple->src.u.all;
    flow->dport = tuple->dst.u.all;

    bpf_probe_read(&flow->saddr, sizeof(flow->saddr), &tuple->src.u3.all);
    bpf_probe_read(&flow->daddr, sizeof(flow->daddr), &tuple->dst.u3.all);
}

struct bpf_map_def SEC("maps/conntrack") conntrack = {
    .type = BPF_MAP_TYPE_LRU_HASH,
    .key_size = sizeof(struct namespaced_flow_t),
    .value_size = sizeof(struct namespaced_flow_t),
    .max_entries = 4096,
    .pinning = 0,
    .namespace = "",
};

__attribute__((always_inline)) int trace_nat_manip_pkt(struct pt_regs *ctx, struct nf_conn *ct) {
    u32 netns = get_netns_from_nf_conn(ct);

    struct nf_conntrack_tuple_hash tuplehash[IP_CT_DIR_MAX];
    bpf_probe_read(&tuplehash, sizeof(tuplehash), &ct->tuplehash);

    struct nf_conntrack_tuple *orig_tuple = &tuplehash[IP_CT_DIR_ORIGINAL].tuple;
    struct nf_conntrack_tuple *reply_tuple = &tuplehash[IP_CT_DIR_REPLY].tuple;

    // parse nat flows
    struct namespaced_flow_t orig = {
        .netns = netns,
    };
    struct namespaced_flow_t reply = {
        .netns = netns,
    };
    parse_tuple(orig_tuple, &orig.flow);
    parse_tuple(reply_tuple, &reply.flow);

    // save nat translation:
    //   - flip(reply) should be mapped to orig
    //   - reply should be mapped to flip(orig)
    flip(&reply.flow);
    bpf_map_update_elem(&conntrack, &reply, &orig, BPF_ANY);
    flip(&reply.flow);
    flip(&orig.flow);
    bpf_map_update_elem(&conntrack, &reply, &orig, BPF_ANY);
    return 0;
}

SEC("kprobe/nf_nat_manip_pkt")
int kprobe_nf_nat_manip_pkt(struct pt_regs *ctx) {
    struct nf_conn *ct = (struct nf_conn *)PT_REGS_PARM2(ctx);
    return trace_nat_manip_pkt(ctx, ct);
}

SEC("kprobe/nf_nat_packet")
int kprobe_nf_nat_packet(struct pt_regs *ctx) {
    struct nf_conn *ct = (struct nf_conn *)PT_REGS_PARM1(ctx);
    return trace_nat_manip_pkt(ctx, ct);
}

#endif
