function check (a, f)
    f = f or function (x,y) return x<y end
    for n=table.getn(a),2,-1 do
        assert(not f(a[n], a[n-1]))
    end
end

