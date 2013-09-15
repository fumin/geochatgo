require "thread"
require "socket"
require "net/http"

Lat = 24.2019
Lng = 120.5832
NearestNeighborsCount = 100

def simulateClient lat: Lat, lng: Lng, log: false
  c = TCPSocket.new 'localhost', 3000
  c.puts "GET /stream HTTP/1.1"
  c.puts "Host: localhost:3000"
  c.puts "\n"

  # Retrive our server generated username
  username = nil
  while true do
    resp = c.gets
    if resp.nil?
      c.close; exit
    else
      puts resp if log
      if resp =~ /event: username/
        resp = c.gets
        if resp.nil?
          c.close; exit
        end
        puts resp if log
        m = /data: ([-_\w]+)/.match(resp)
        if m[1].nil?
          puts "ERROR: #{resp}"; c.close; Thread.current.exit
        else
          username = m[1]
          break
        end
      end
    end
  end

  # Simulate update map bounds
  Net::HTTP.post_form(URI("http://localhost:3000/update_mapbounds"),
                      username: username, south: lat+0.0001, west: lng+0.0001,
                      north: 30, east: 130)

  while true do
    resp = c.gets
    if resp.nil?
      c.close
      break
    else
      puts resp if log
    end
  end
end

# ---------------------------------------------
# Start main function
puts "The test is as follows:"
puts "Open browser and open the Javascipt console."
puts "If our map's bounds include 台中, we'll receive what we type here."
puts "Conversely, if we are out of 台中, we won't receive anything."

(NearestNeighborsCount-1).times{
  Thread.new{ simulateClient }
}
Thread.new{ simulateClient(lat: Lat+0.1, lng: Lng+0.1, log: true) }

while true do
  puts "----------------------"
  print "Please say something:"
  msg = gets
  puts Net::HTTP.post_form(URI("http://localhost:3000/say"),
                           msg: msg, latitude: Lat, longitude: Lng).body
end
